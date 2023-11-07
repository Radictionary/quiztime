package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Radictionary/kahoot/backend/internals/game"
	"github.com/Radictionary/kahoot/backend/internals/models"
	"github.com/Radictionary/kahoot/backend/internals/render"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

func (m *Repository) LoginRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Println("error parsing form:", err)
		return
	}
	userName := strings.ToLower(r.Form.Get("name"))
	creds := models.Account{
		Name:     userName,
		Password: r.Form.Get("password"),
	}
	storedCreds, err := m.App.Redis.RetrieveUserAccount(userName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Println("user doesn't exist:", err)
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Println("Couldn't retrieve stored database credentials: ", err)
		return
	}
	// Compare the stored hashed password, with the password that was received
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		fmt.Println(storedCreds, "\n", creds)
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}

	sessionToken := uuid.NewString()
	current_time := time.Now()
	storedCreds.UserStatistics.LastLoggedIn = current_time.Format("03:04 PM 01/02/2006")
	expiresAt := time.Now().Add(15 * time.Minute)
	sessionData, err := json.Marshal(storedCreds)
	if err != nil {
		log.Println("error marshaling data:", err)
	}
	m.App.Session.Store.Commit(
		sessionToken,
		sessionData,
		expiresAt,
	)
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionToken",
		Value:    sessionToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.App.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	redirectPage := m.App.Session.PopString(r.Context(), "attempted_page")
	if redirectPage == "" {
		redirectPage = "/"
	}

	m.App.Redis.StoreUserAccount(storedCreds)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(redirectPage)) //Tell frontend where to redirect to next
}

func (m *Repository) SignupRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Println("error parsing form:", err)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), 10)
	if err != nil {
		log.Println("error bcrypt hashing password:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	form := models.Account{
		Name:     strings.ToLower(r.FormValue("name")),
		Password: string(hashedPassword),
	}
	userCheck, _ := m.App.Redis.RetrieveUserAccount(r.FormValue("name"))
	if userCheck.Name != "" {
		http.Error(w, "name taken", http.StatusForbidden)
		return
	}
	err = m.App.Redis.StoreUserAccount(form)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Println("error storing user account:", err)
		return
	}
	m.App.Session.Put(r.Context(), "flash", "Account Created")
	w.WriteHeader(http.StatusOK)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := getSessionData(r)
	if !loggedIn {
		m.App.Session.Put(r.Context(), "flash", "You are not logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionToken",
		Value:    "",
		Expires:  time.Now(),
		Path:     "/",
		HttpOnly: true,
		Secure:   m.App.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	m.App.Session.Put(r.Context(), "flash", "Successfully logged out")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (m *Repository) Accounts(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error reading body:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	oldName := chi.URLParam(r, "name")
	oldAccount, err := m.App.Redis.RetrieveUserAccount(oldName)
	if err != nil {
		fmt.Println("Error retriving map from redis db:", err)
	}
	sessionData, _ := getSessionData(r)
	if sessionData.Name != oldName {
		m.App.Session.Put(r.Context(), "flash", "You are not authorized")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if sessionData.Password != oldAccount.Password {
		m.App.Session.Put(r.Context(), "flash", "You are not authorized")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	newAccount := models.Account{}
	err = json.Unmarshal(body, &newAccount)
	if err != nil {
		http.Error(w, "Failed to unmarshal JSON data", http.StatusBadRequest)
		return
	}
	newAccount.Password = oldAccount.Password
	m.App.Redis.Remove("account:" + oldName)
	m.App.Redis.StoreUserAccount(newAccount)
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(15 * time.Minute)
	sessionDataJSON, err := json.Marshal(newAccount)
	if err != nil {
		fmt.Println("error marshaling data:", err)
	}
	m.App.Session.Store.Commit(
		sessionToken,
		sessionDataJSON,
		expiresAt,
	)
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionToken",
		Value:    sessionToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.App.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *Repository) Game(w http.ResponseWriter, r *http.Request) {
	account, loggedIn := getSessionData(r)
	gameName := chi.URLParam(r, "name")

	switch r.Method {
	case "GET": //serving html
		currentAccount, _ := m.App.Redis.RetrieveUserAccount(account.Name)
		if !currentAccount.AllowedGame(gameName) {
			m.App.Session.Put(r.Context(), "flash", "You don't own this game!")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		game, _ := m.App.Redis.RetrieveGame(gameName)
		if strings.Contains(r.URL.String(), "/startgame") {
			test := m.App.Game.NewGame(game)
			fmt.Println(test.GameCode)
			w.Write([]byte(strconv.Itoa(test.GameCode)))
			return
		}
		gameJSON, _ := json.Marshal(game)
		fmt.Println("the game retrieved when loading gamesetup is:", game)
		render.RenderTemplate(w, "gamesetup.html", &models.TemplateData{
			LoggedIn:       loggedIn,
			Account:        currentAccount,
			Message:        gameName,
			GameJSON:       string(gameJSON),
			ProfilePicture: m.getProfilePicture(currentAccount.Name),
		})
	case "POST":
		currentAccount, _ := m.App.Redis.RetrieveUserAccount(account.Name)

		if gameCheck, _ := m.App.Redis.RetrieveGame(gameName); gameCheck.DateOfCreation != "" {
			http.Error(w, "game name taken", http.StatusForbidden)
			return
		}

		currentAccount.Games = append(currentAccount.Games, gameName)
		m.App.Redis.StoreUserAccount(currentAccount)
		w.WriteHeader(http.StatusOK)
	case "PUT":
		currentAccount, _ := m.App.Redis.RetrieveUserAccount(account.Name)
		if !currentAccount.AllowedGame(gameName) {
			m.App.Session.Put(r.Context(), "flash", "You don't own this game!")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if strings.Contains(r.URL.String(), "/share") {
			userToShare := r.URL.Query().Get("user")
			userToShareAccount, _ := m.App.Redis.RetrieveUserAccount(userToShare)
			userToShareAccount.SharedGames = append(userToShareAccount.SharedGames, gameName)
			m.App.Redis.StoreUserAccount(userToShareAccount)
			game, _ := m.App.Redis.RetrieveGame(gameName)
			game.UsersShared = append(game.UsersShared, userToShare)
			m.App.Redis.StoreGame(game)
			w.WriteHeader(http.StatusOK)
		} else {
			oldGame, _ := m.App.Redis.RetrieveGame(gameName)


			var game game.Game
			if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
				log.Println("error decoding body when creating/editing game:", err)
				return
			}
			if gameTest, _ := m.App.Redis.RetrieveGame(game.Name); gameTest.DateOfCreation == "" { //only update if it is first time creating game
				game.DateOfCreation = time.Now().Format(time.UnixDate)
			} else{
				game.DateOfCreation = oldGame.DateOfCreation
			}
			m.App.Redis.StoreGame(game)
			w.WriteHeader(http.StatusOK)
		}

	case "DELETE":
		gameName := chi.URLParam(r, "name")
		if strings.Contains(r.URL.String(), "/share") {
			userToShare := r.URL.Query().Get("user")

			userToShareAccount, _ := m.App.Redis.RetrieveUserAccount(userToShare)
			userToShareAccount.SharedGames = models.RemoveStringFromStringSlice(userToShareAccount.SharedGames, gameName)
			m.App.Redis.StoreUserAccount(userToShareAccount)

			game, _ := m.App.Redis.RetrieveGame(gameName)

			game.UsersShared = models.RemoveStringFromStringSlice(game.UsersShared, userToShare)
			err := m.App.Redis.StoreGame(game)
			if err != nil {
				fmt.Println("error is:", err)
			}
			w.WriteHeader(http.StatusOK)

		} else {
			game, _ := m.App.Redis.RetrieveGame(gameName)
			for _, v := range game.UsersShared {
				user, _ := m.App.Redis.RetrieveUserAccount(v)
				user.SharedGames = models.RemoveStringFromStringSlice(user.SharedGames, gameName)
				m.App.Redis.StoreUserAccount(user)
			}
			gameOwner, _ := m.App.Redis.RetrieveUserAccount(game.Owner)
			gameOwner.Games = models.RemoveStringFromStringSlice(gameOwner.Games, gameName)
			m.App.Redis.StoreUserAccount(gameOwner)

			m.App.Redis.Remove("game:" + game.Name)
			w.WriteHeader(http.StatusOK)
		}
	}
}


var upgrader = websocket.Upgrader{
	CheckOrigin: websocket.IsWebSocketUpgrade,
} // use default options
func (m *Repository) PlayGameWS(w http.ResponseWriter, r *http.Request) { //TODO implement disconnect handler
	w.Header().Set("Access-Control-Allow-Origin", "*")
	name := r.URL.Query().Get("name")
	picture := r.URL.Query().Get("picture")
	picture, _ = url.PathUnescape(picture)
	gameName := chi.URLParam(r, "code")
	gameCode, _ := strconv.Atoi(gameName)
	currentGame, found := m.App.Game.Games[gameCode]
	if !found {
		http.Redirect(w, r, "/join", http.StatusNotFound)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error upgrading to websocket conn:", err)
		http.Error(w, "error upgrading to websocket conn:"+err.Error(), http.StatusInternalServerError)
	}
	player := game.Player{
		Name:        name,
		Playing:     true, //TODO: make it so that a user can choose to spectate and not need to play
		Conn:        conn,
		Picture:     picture,
		SendChan:    make(chan game.ClientMessage),
		ReceiveChan: make(chan game.ClientMessage),
	}
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("player left")
		player.Conn = nil
		currentGame.Players[player.Name] = &player
		currentGame.RemovePlayer(player.Name, "") //reason is blank("") as there is no point of displaying a reason to a absent player
		return nil
	})
	go func() { // start goroutine which listens to channel for messages to send
		for {
			message := <-player.SendChan
			conn.WriteJSON(message)
		}
	}()
	if len(currentGame.Players) == 0 {
		player.Owner = true
		currentGame.Players[player.Name] = &player
	} else {
		currentGame.Players[player.Name] = &player
		currentGame.FindOwner().SendJSON("connectedUsers", currentGame.ListUsers()) //send to owner the list of users connected to game. If the user connected is the owner, there is no need
	}
	go player.StartGameListening(currentGame) //processes received messages
	for player.Conn != nil { //start goroutine which listens for incoming messages and sends them to receive channel for processing
		_, rawMessage, err := player.Conn.ReadMessage()
		if err != nil {
			fmt.Println("ERROR READING MESSAGE FROM", player.Name, "AND ERROR IS:", err)
			return
		}
		var message game.ClientMessage
		json.Unmarshal(rawMessage, &message)
		player.ReceiveChan <- message
	}
}