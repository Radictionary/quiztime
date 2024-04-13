package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Radictionary/kahoot/internals/models"
	"github.com/google/uuid"
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
		//fmt.Println(storedCreds, "\n", creds)
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(15 * time.Minute)
	sessionData, err := json.Marshal(storedCreds)
	if err != nil {
		log.Println("error marshaling data:", err)
		return
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
		redirectPage = "/dashboard"
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
		http.Error(w, "name taken", http.StatusConflict)
		return
	}
	err = m.App.Redis.StoreUserAccount(form)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Println("error storing user account:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := getSessionData(r)
	if !loggedIn {
		http.Redirect(w, r, "/?message=You+are+not+logged+in", http.StatusSeeOther)
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
	http.Redirect(w, r, "/login?message=Successfully+logged+out", http.StatusSeeOther)
}
