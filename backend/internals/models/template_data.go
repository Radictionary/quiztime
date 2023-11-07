package models

import (
	"html/template"

	"github.com/Radictionary/kahoot/backend/internals/game"
)

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	StringMap map[string]string
	DataMap      map[string]any
	BoolMap map[string]bool
	CSRFToken string //Cross Site Request Forgery Token
	Message string
	Flash     string
	LoggedIn bool
	Account Account
	AccountJSON string //string([]byte)
	Games []game.Game
	SharedGames []game.Game
	SharedUsersJSON string //string([]byte)
	Game game.Game
	GameJSON string //string([]byte)
	ProfilePicture template.HTML
	GameAdmin bool
	Test template.HTML
}
