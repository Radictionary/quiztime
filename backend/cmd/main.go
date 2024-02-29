package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Radictionary/kahoot/backend/cmd/handlers"
	"github.com/Radictionary/kahoot/backend/internals/config"
	"github.com/Radictionary/kahoot/backend/internals/game"
	"github.com/Radictionary/kahoot/backend/internals/redis"
	"github.com/Radictionary/kahoot/backend/internals/render"
	"github.com/alexedwards/scs/v2"
)

const (
	portNumber       = ":8082"
	backupPortNumber = ":7083"
)

var (
	app     config.AppConfig
	session *scs.SessionManager
)

// main is the main function
func main() {
	//change this to true when in production
	app.InProduction = true

	session = scs.New()
	session.Lifetime = 2 * time.Hour
	session.Cookie.Persist = app.InProduction
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		fmt.Println("cannot create template cache:", err)
		panic("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction

	app.Redis = redis.InitRedisConnection()

	app.Game = game.InitGamesHub()

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)

	fmt.Printf("Staring application on port %s\n", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Printf("Could not start application using port number %s, using fallback port number %s\n", portNumber, backupPortNumber)
		srv.Addr = backupPortNumber
		fmt.Println("Setting production to false")
		app.InProduction = false
		err = srv.ListenAndServe()
		if err != nil {
			fmt.Println("Could not start web application")
			return
		}
	}
	log.Println("Could not start web application")
	log.Fatal(err)
}
