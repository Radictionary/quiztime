package config

import (
	"html/template"

	"github.com/Radictionary/kahoot/backend/internals/game"
	"github.com/Radictionary/kahoot/backend/internals/redis"
	"github.com/alexedwards/scs/v2"
)

// AppConfig holds the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InProduction bool
	Session *scs.SessionManager
	Redis *redis.RedisConn
	Game *game.GamesHub
}
