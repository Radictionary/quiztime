package models

import "fmt"

type Account struct {
	Name           string         `json:"name"`
	ProfilePicture string         `json:"profilePicture"`
	Password       string         `json:"password"`
	Games          []string       `json:"games"`
	SharedGames    []string       `json:"sharedGames"`
	UserStatistics UserStatistics `json:"statistics"`
}

type UserStatistics struct {
	LastLoggedIn string `json:"lastLoggedIn"`
}

func (a *Account) AllowedGame(gameName string) bool {
	for _, game := range a.Games {
		if game == gameName { //user owns this game
			return true

		}
	}
	for _, game := range a.SharedGames {
		if game == gameName {
			return true
		}
	}
	fmt.Println("declined")
	return false
}
