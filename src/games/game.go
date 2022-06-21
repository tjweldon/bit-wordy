package games

import (
	"bit-wordy/src/primitives"
	"fmt"
)

// Game represents an instance of a wordle game including its
// internal game state
type Game struct {
	Answer  primitives.Fivegram
	Results primitives.ResultSet
}

// NewGame returns a fresh game with the answer passed
func NewGame(answer primitives.Fivegram) *Game {
	return &Game{
		Answer:  answer,
		Results: primitives.ResultSet{},
	}
}

// String
func (g Game) String() string {
	return fmt.Sprintf("Answer: %s\n%s\n", g.Answer, g.Results)
}

// Guess is the function corresponding to a single attempt to guess
// the answer
func (g *Game) Guess(word primitives.Fivegram) primitives.Pattern {
	pattern := word.CheckGuess(g.Answer)

	result := primitives.Result{Word: word, Pattern: pattern}
	g.Results = append(g.Results, result)

	return pattern
}

// IsWon returns true if the latest guess was a winner
func (g Game) IsWon() bool {
	return g.Results[len(g.Results)-1].Word == g.Answer
}

// IsLost returns true if the number of guesses gets to 6 and the answer is not found
func (g Game) IsLost() bool {
	return len(g.Results) >= 6 && !g.IsWon()
}
