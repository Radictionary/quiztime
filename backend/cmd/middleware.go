package main

import (
	"context"
	"log"
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/realtime/admin" || r.URL.Path == "/status" {
			next.ServeHTTP(w, r)
			return
		} else {
			session.LoadAndSave(next).ServeHTTP(w, r)
		}
	})
}


// ValidateSession ensures the route is accessed by a logged in user with valid session
func ValidateSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptedPage := r.URL
		c, err := r.Cookie("sessionToken")
		if err != nil {
			if err == http.ErrNoCookie {
				app.Session.Put(r.Context(), "flash", "Log in to continue")
				
				app.Session.Put(r.Context(), "attempted_page", attemptedPage.String())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				w.WriteHeader(http.StatusUnauthorized)
				return
			} 
			log.Println("Error retrieving cookie:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sessionDataJSON, found, err := app.Session.Store.Find(c.Value)
		if err != nil {
			log.Println("Err:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !found {
			app.Session.Put(r.Context(), "flash", "Invalid Session")
			app.Session.Put(r.Context(), "attempted_page", attemptedPage.String())
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Store the session data in the request context
		ctx := context.WithValue(r.Context(), "sessionData", sessionDataJSON)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// SoftValidateSession checks to see if the user is logged in to customize the experience and gather session data. Does not require an active session
func SoftValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessionDataJSON []byte

		c, err := r.Cookie("sessionToken")
		if err != nil || c == nil {
			sessionDataJSON = nil
		} else {
			var found bool
			sessionDataJSON, found, err = app.Session.Store.Find(c.Value)
			if err != nil {
				log.Println("Error finding sessionData from middleware:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !found {
				sessionDataJSON = nil
			}
		}
		// Store the session data in the request context
		ctx := context.WithValue(r.Context(), "sessionData", sessionDataJSON)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
