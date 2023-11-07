package main

import (
	"encoding/json"
	"net/http"

	"github.com/Radictionary/kahoot/backend/cmd/handlers"
	"github.com/Radictionary/kahoot/backend/internals/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(SoftValidateSession)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	//Non websocket connections
	r.Group(func(r chi.Router) {
		r.Use(SessionLoad) //Session Load Breaks Websocket connections
		//IMPLEMENT CSRF
		r.Get("/", ValidateSession(handlers.Repo.Home))

		r.Post("/loginrequest", handlers.Repo.LoginRequest)
		r.Post("/signuprequest", handlers.Repo.SignupRequest)
		r.Get("/logout", handlers.Repo.Logout)

		r.Put("/accounts/{name}", handlers.Repo.Accounts)

		r.Route("/game/{name}", func(r chi.Router) {
			r.Get("/", handlers.Repo.Game)
			r.Post("/", handlers.Repo.Game)
			r.Put("/", handlers.Repo.Game)
			r.Delete("/", handlers.Repo.Game)
			r.Put("/share", handlers.Repo.Game)
			r.Delete("/share", handlers.Repo.Game)
			r.Get("/startgame", handlers.Repo.Game)
		})

		r.Get("/play/{code}", handlers.Repo.PlayGame)
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			test, _ := json.Marshal(handlers.Repo.App.Game.Games)
			w.Write([]byte(string(test)))
		})

		r.Get("/login", handlers.Repo.Login)
		r.Get("/signup", handlers.Repo.Signup)
		r.Get("/profile", ValidateSession(handlers.Repo.Profile))
		r.Get("/join", handlers.Repo.JoinGame)
	})
	r.Get("/play/{code}/ws", handlers.Repo.PlayGameWS)

	fileServer := http.FileServer(http.Dir("../../frontend/"))
	r.Handle("/frontend/*", http.StripPrefix("/frontend", fileServer))

	return r
}
