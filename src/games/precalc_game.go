package games

import (
	"bit-wordy/src/primitives"
	"bit-wordy/src/splat"
	"fmt"
)

// FatGame represents an instance of a wordle game including its
// internal game state. It uses an in-memory byte slice as a cache of
// word match patterns
type FatGame struct {
	splt    *splat.Splat
	Answer  primitives.Fivegram
	Results primitives.ResultSet
}

// NewFatGame returns a fresh game with the answer passed
func NewFatGame(answer primitives.Fivegram, splat *splat.Splat) *FatGame {
	return &FatGame{
		splt:    splat,
		Answer:  answer,
		Results: primitives.ResultSet{},
	}
}

// String
func (g FatGame) String() string {
	return fmt.Sprintf("Answer: %s\n%s\n", g.Answer, g.Results)
}

// Guess is the function corresponding to a single attempt to guess
// the answer
func (g *FatGame) Guess(word primitives.Fivegram) primitives.Pattern {
	pattern := word.CheckGuess(g.Answer)

	result := primitives.Result{Word: word, Pattern: pattern}
	g.Results = append(g.Results, result)

	return pattern
}

// IsWon returns true if the latest guess was a winner
func (g FatGame) IsWon() bool {
	return g.Results[len(g.Results)-1].Word == g.Answer
}

// IsLost returns true if the number of guesses gets to 6 and the answer is not found
func (g FatGame) IsLost() bool {
	return len(g.Results) >= 6 && !g.IsWon()
}
