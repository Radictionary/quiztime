package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

//Map of map[gameCode]*Game
type GamesHub struct {
	Games map[int]*Game
}

//Game struct
type Game struct {
	Name                     string
	Owner                    string
	UsersShared              []string
	Status                   string
	Timer                    int //in seconds
	Questions                int
	QuestionsAnswers         []QuestionsAnswers
	Players                  map[string]*Player //When game is being played. First person to join game is the game instance owner. Assumed this will always be the person who starts the game
	CurrentQuestion          int
	DateOfCreation           string
	GamePicture              string
	GameCode                 int
	QuestionsAnsweredInRound int
}

//Each Question and Answer, as well as the information needed about them
type QuestionsAnswers struct {
	QuestionNumber int
	Question       string
	Answers        []string //includes correct and wrong answers
	CorrectAnswers []string
	Points         int
	QuestionImage  string //needs implementation, if at all
}

//Each Player and the information needed about them
type Player struct {
	Name            string
	Owner           bool //for when game is being played. Is the first person to join game.
	Conn            *websocket.Conn
	Points          int
	answeredCorrect int
	answeredWrong   int
	Playing         bool
	AnsweredCorrect bool //for logic which will will loop through all players and tell them if they answered correct. Value constantly changing. Might not use
	Picture         string
	SendChan        chan ClientMessage //chan to send to frontend concurrently. Any message sent to the frontend must be sent through this channel
	ReceiveChan     chan ClientMessage //chan to receive frontend messages
}

//Type used to communicate backend and client
type ClientMessage struct {
	Label   string `json:"label"`
	Message string `json:"message"` //json: golang implements with binary, frontend javascript uses strings; will be string
}

//Initialize the gameHub
func InitGamesHub() *GamesHub {
	return &GamesHub{
		Games: make(map[int]*Game),
	}
}

//Create a new game, it's game code, and add it to the gameHub registry
func (h *GamesHub) NewGame(game Game) *Game {
	var gameCode int
	for {
		gameCode = rand.Intn(1000000)
		if _, found := h.Games[gameCode]; !found {
			break
		}
	}
	game.Players = make(map[string]*Player)
	game.Status = "waiting"
	game.GameCode = gameCode
	h.Games[gameCode] = &game

	return &game
}

//Find game owner, if none found, will return empty Player with the name of "NO OWNER FOUND"
func (g *Game) FindOwner() *Player {
	for _, player := range g.Players {
		if player.Owner {
			return player
		}
	}
	fmt.Println("no owner found")
	return &Player{Name: "NO OWNER FOUND"}
}

//Returns all users in the game with their name, points, and picture in the form of a map[string]string
func (g *Game) ListUsers() []map[string]string {
	var players []map[string]string
	for _, player := range g.Players {
		players = append(players, map[string]string{"name": player.Name, "picture": player.Picture})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i]["name"] < players[j]["name"]
	})
	return players
}

//If answer is correct, give the player points based on the kahoot algorithm
func (g *Game) ProcessCorrectPlayerAnswer(p *Player, timerWhenAnswered int) { //points algorithm at https://support.kahoot.com/hc/en-us/articles/115002303908-How-points-work
	questionsAnswers := g.QuestionsAnswers[g.CurrentQuestion-1]
	test := p.Points
	p.Points += int((1 - (((g.Timer - timerWhenAnswered) / g.Timer) / 2)) * questionsAnswers.Points)
	fmt.Printf("gave %v: %v points", p.Name, p.Points - test)
}

//Removes a player from the game
func (g *Game) RemovePlayer(name string, reason string) { //TODO: make function more optimized by using generics and allowing pointer to player as parameter also
	player, found := g.Players[name]
	if !found {
		fmt.Println("didn't find user to remove:", name)
	}
	if player.Conn != nil && !player.Owner { //game owner can't be kicked
		player.SendMessage("leave", reason)

		<-time.After(time.Second) //wait for frontend client attempt to leave. If not done after one second, forcefully close the connection
		// player.Conn.Close() //causes panics
		player.Conn = nil
	}
	if player.Owner {
		for _, player := range g.Players {
			if !player.Owner {
				g.RemovePlayer(player.Name, "owner left")
			}
		}
	}
	delete(g.Players, name)
}

//Generates and sends scoreboard to the players in the game
func (g *Game) SendScoreboard(endGame bool) {
	var playersSlice []Player
	for _, player := range g.Players {
		if !player.Owner {
			playersSlice = append(playersSlice, *player)
		}
	}

	// Sort the playerSlice in descending order of points
	sort.Slice(playersSlice, func(i, j int) bool {
		return playersSlice[i].Points > playersSlice[j].Points
	})
	if len(playersSlice) > 5 {
		playersSlice = playersSlice[:5]
	}
	var playerSliceFormatted []map[string]string
	for _, player := range playersSlice {
		playerSliceFormatted = append(playerSliceFormatted, map[string]string{
			"name":    player.Name,
			"picture": player.Picture,
			"points":  strconv.Itoa(player.Points),
		})
	}
	if endGame {
		for _, player := range g.Players {
			player.SendJSON("endgame", playerSliceFormatted)
		}
	} else {
		for _, player := range g.Players {
			player.SendJSON("scoreboard", playerSliceFormatted)
		}
	}

}

//Starts listening for client messages for the backend to process
func (p *Player) StartGameListening(g *Game) {
	for {
		clientMessage := <-p.ReceiveChan
		fmt.Println("received message:", clientMessage)
		switch string(clientMessage.Label) {
		case "startGame":
			if !p.Owner { //owner only
				return
			}
			g.Status = "started"
			g.CurrentQuestion = 1
			for _, player := range g.Players {
				neededQuestion := g.QuestionsAnswers[g.CurrentQuestion-1]
				player.SendJSON("question", neededQuestion)
			}
		case "kick":
			if !p.Owner || (clientMessage.Message == p.Name && p.Owner) { //owner only and owner can't be kicked from their own game
				return
			}
			g.RemovePlayer(clientMessage.Message, "you are kicked") //displayed to frontend as "you are leaving because 'you are kicked'"
		case "question":
			if g.CurrentQuestion > len(g.QuestionsAnswers) {
				g.SendScoreboard(true)
			} else {
				for _, player := range g.Players {
					neededQuestion := g.QuestionsAnswers[g.CurrentQuestion-1]
					player.SendJSON("question", neededQuestion)
				}
				// neededQuestion := g.QuestionsAnswers[g.CurrentQuestion-1]
				// p.SendJSON("question", neededQuestion)
			}
		case "answer":
			var userAnswer map[string]string
			json.Unmarshal([]byte(clientMessage.Message), &userAnswer)
			g.QuestionsAnsweredInRound++
			currentQuestion := g.QuestionsAnswers[g.CurrentQuestion-1]
			// fmt.Println("third:", g.QuestionsAnswers[g.CurrentQuestion].CorrectAnswers[len(g.QuestionsAnswers[g.CurrentQuestion].CorrectAnswers)-1])
			if len(currentQuestion.CorrectAnswers) > 1 {
				for _, v := range currentQuestion.CorrectAnswers {
					if v == userAnswer["answer"] {
						timerWhenAnswered, _ := strconv.Atoi(userAnswer["timerWhenAnswered"])
						g.ProcessCorrectPlayerAnswer(p, timerWhenAnswered)
					}
				}
			} else {
				if userAnswer["answer"] == currentQuestion.CorrectAnswers[0] {
					timerWhenAnswered, _ := strconv.Atoi(userAnswer["timerWhenAnswered"])
					g.ProcessCorrectPlayerAnswer(p, timerWhenAnswered)
				}
			}
			if g.QuestionsAnsweredInRound >= len(g.Players)-1 {
				g.CurrentQuestion++
				g.QuestionsAnsweredInRound = 0
				g.SendScoreboard(false)
			}
			// case "outOfTime":
			// 	g.CurrentQuestion++
			// 	g.SendScoreboard(false)
		}
	}
}

//Sends JSON(Javascript Object Notation) to the frontend with the ClientMessage type
func (p *Player) SendJSON(label string, v any) {
	jsonData, _ := json.Marshal(v)
	p.SendChan <- ClientMessage{
		Label:   label,
		Message: string(jsonData),
	}
}

//Sends a string message to the frontend with the ClientMessage type
func (p *Player) SendMessage(label string, s string) {
	p.SendChan <- ClientMessage{
		Label:   label,
		Message: s,
	}
}
