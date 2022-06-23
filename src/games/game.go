package games

import (
	"bit-wordy/src/primitives"
	"fmt"
)

// Game represents an instance of a wordle game including its
// internal game state
type Game struct {
	Answer     primitives.Word
	Results    primitives.ResultSet
	CheckGuess func(guess, ans primitives.Word) primitives.Pattern
}

// NewGame returns a fresh game with the answer passed
func NewGame(answer primitives.Word) *Game {
	return &Game{
		Answer:  answer,
		Results: primitives.ResultSet{},
	}
}

// String
func (g Game) String() string {
	outcome := "In Progress"
	if g.IsWon() {
		outcome = "Won!"
	} else if g.IsLost() {
		outcome = "Lost :("
	}
	return fmt.Sprintf("Answer: %s\n%s\nOutcome: %s\n", g.Answer, g.Results, outcome)
}

// Guess is the function corresponding to a single attempt to guess
// the answer
func (g *Game) Guess(word primitives.Word) primitives.Pattern {
	pattern := g.Answer.CheckGuess(word)

	result := primitives.Result{Word: word, Pattern: pattern}
	g.Results = append(g.Results, result)

	return pattern
}

// IsWon returns true if the latest guess was a winner
func (g Game) IsWon() bool {
	isWon := false
	if len(g.Results) >= 1 {
		isWon = g.Results[len(g.Results)-1].Word == g.Answer
	}

	return isWon
}

// IsLost returns true if the number of guesses gets to 6 and the answer is not found
func (g Game) IsLost() bool {
	isLost := len(g.Results) >= 6 && !g.IsWon()

	return isLost
}

func (g *Game) Reset(ans primitives.Word) {
	g.Results = primitives.ResultSet{}
	g.Answer = ans
}
