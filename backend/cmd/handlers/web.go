package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Radictionary/kahoot/backend/internals/config"
	"github.com/Radictionary/kahoot/backend/internals/game"
	"github.com/Radictionary/kahoot/backend/internals/models"
	"github.com/Radictionary/kahoot/backend/internals/render"
	"github.com/go-chi/chi/v5"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	account, loggedIn := getSessionData(r)
	games := m.FindGames(&account)

	sharedGames := m.FindSharedGames(&account)
	profilePicture := m.getProfilePicture(account.Name)

	render.RenderTemplate(w, "index.html", &models.TemplateData{
		LoggedIn:    loggedIn,
		Account:     account,
		Games:       games,
		SharedGames: sharedGames,
		ProfilePicture: profilePicture,
	})
}
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := getSessionData(r)
	flash, _ := m.App.Session.Pop(r.Context(), "flash").(string) //global flash

	if loggedIn {
		m.App.Session.Put(r.Context(), "flash", "Already Logged In")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	render.RenderTemplate(w, "login.html", &models.TemplateData{
		Flash: flash,
	})
}
func (m *Repository) Signup(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := getSessionData(r)
	if loggedIn {
		m.App.Session.Put(r.Context(), "flash", "Already Logged In, Already Created An Account")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	render.RenderTemplate(w, "signup.html", &models.TemplateData{})
}

func (m *Repository) Profile(w http.ResponseWriter, r *http.Request) {
	account, loggedIn := getSessionData(r)

	render.RenderTemplate(w, "profile.html", &models.TemplateData{
		LoggedIn:       loggedIn,
		Account:        account,
		AccountJSON:    string(m.getAccountJSON(account.Name)),
		ProfilePicture: m.getProfilePicture(account.Name),
	})
}

func (m *Repository) JoinGame(w http.ResponseWriter, r *http.Request) {
	account, loggedIn := getSessionData(r)
	render.RenderTemplate(w, "join.html", &models.TemplateData{
		LoggedIn:       loggedIn,
		Account:        account,
		AccountJSON:    string(m.getAccountJSON(account.Name)),
		ProfilePicture: m.getProfilePicture(account.Name),
	})
}

func (m *Repository) PlayGame(w http.ResponseWriter, r *http.Request) {
	account, loggedIn := getSessionData(r)
	gameCodeString := chi.URLParam(r, "code")
	gameCode, _ := strconv.Atoi(gameCodeString)
	game, found := m.App.Game.Games[gameCode]
	if !found {
		http.Redirect(w, r, "/join", http.StatusNotFound)
		return
	}
	var gameAdmin bool
	if len(game.Players) == 0 {
		gameAdmin = true
	}
	if game.Status != "waiting" {
		http.Redirect(w, r, "/join", http.StatusForbidden)
		return
	}
	render.RenderTemplate(w, "game.html", &models.TemplateData{
		LoggedIn:       loggedIn,
		Account:        account,
		ProfilePicture: m.getProfilePicture(account.Name),
		AccountJSON:    string(m.getAccountJSON(account.Name)),
		Game:           *game,
		GameAdmin:      gameAdmin,
	})
}

func (m *Repository) FindGames(a *models.Account) []game.Game {
	var games []game.Game
	account, _ := m.App.Redis.RetrieveUserAccount(a.Name)
	for _, game := range account.Games {
		game, _ := m.App.Redis.RetrieveGame(game) //TODO HANDLE ALL REDIS RETRIEVAL ERROR
		games = append(games, game)
	}
	return games
}

func (m *Repository) FindSharedGames(a *models.Account) []game.Game {
	var sharedGames []game.Game
	account, _ := m.App.Redis.RetrieveUserAccount(a.Name)
	for _, game := range account.SharedGames {
		game, _ := m.App.Redis.RetrieveGame(game) //TODO HANDLE ALL REDIS RETRIEVAL ERROR
		sharedGames = append(sharedGames, game)
	}
	return sharedGames
}

func getSessionData(r *http.Request) (models.Account, bool) {
	var sessionData models.Account
	sessionDataJSON, _ := r.Context().Value("sessionData").([]byte)
	if sessionDataJSON == nil {
		return models.Account{}, false
	}
	err := json.Unmarshal(sessionDataJSON, &sessionData)
	if err != nil && sessionDataJSON != nil {
		return models.Account{}, false
	}
	return sessionData, true
}

func (m *Repository) getAccountJSON(name string) []byte {
	userAccount, _ := m.App.Redis.RetrieveUserAccount(name)
	userAccountJSON, _ := json.Marshal(userAccount)
	return userAccountJSON
}

func (m *Repository) getProfilePicture(name string) template.HTML {
	user, _ := m.App.Redis.RetrieveUserAccount(name)
	if user.ProfilePicture == "" {

	}
	return template.HTML(`<img src="data:image/jpeg;base64,` + user.ProfilePicture + `" class="rounded-full h-[30px] w-[30px]" alt="" loading="lazy" id="navbarProfile"/>`) //send as html(no js needed-faster)
}