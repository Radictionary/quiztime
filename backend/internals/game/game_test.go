package game

import "testing"

func TestProcessCorrectPlayerAnswer(t *testing.T) {
    tests := []struct {
        name                string
        playerPoints        int
        timerWhenAnswered   int
        gameTimer           int
        questionPoints      int
        expectedPlayerPoints int
    }{
        {
            name:                "Player answers immediately",
            playerPoints:        0,
            timerWhenAnswered:   0,
            gameTimer:           30,
            questionPoints:      1000,
            expectedPlayerPoints: 1000,
        },
        {
            name:                "Player answers halfway through timer",
            playerPoints:        0,
            timerWhenAnswered:   15,
            gameTimer:           30,
            questionPoints:      1000,
            expectedPlayerPoints: 750,
        },
        {
            name:                "Player answers at the end of timer",
            playerPoints:        0,
            timerWhenAnswered:   30,
            gameTimer:           30,
            questionPoints:      1000,
            expectedPlayerPoints: 500,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            g := &Game{
                Timer: tt.gameTimer,
                QuestionsAnswers: []QuestionsAnswers{
                    {Points: tt.questionPoints},
                },
                CurrentQuestion: 1,
            }
            p := &Player{Points: tt.playerPoints}

            g.ProcessCorrectPlayerAnswer(p, tt.timerWhenAnswered)

            if p.Points != tt.expectedPlayerPoints {
                t.Errorf("ProcessCorrectPlayerAnswer() = %v, want %v", p.Points, tt.expectedPlayerPoints)
            }
        })
    }
}