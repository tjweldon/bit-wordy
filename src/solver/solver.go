package solver

import (
	"bit-wordy/src/games"
	"bit-wordy/src/primitives"
	"bit-wordy/src/util"
	"fmt"
	"log"
	"math"
	"sort"
)

// Solver is a struct that encapsulates the solving algorithm
type Solver struct {
	possibleAnswers primitives.Dictionary
	Game            *games.Game
	Result          primitives.Word
}

// NewSolver returns a reference to a new solver instance
func NewSolver(game *games.Game, dict primitives.Dictionary) *Solver {
	return &Solver{
		possibleAnswers: dict,
		Game:            game,
	}
}

// String
func (s Solver) String() string {
	return fmt.Sprintf("Solver Claims win with:\n%s\n\nGAME:\n%s", s.Result, s.Game)
}

func (s Solver) GameScore() int {
	return len(s.Game.Results)
}

// guessScore exists to simplify guess ranking logic
type guessScore struct {
	guess primitives.Word
	score float64
}

func (s *Solver) Solve() *games.Game {
	var remainingAnswers primitives.Dictionary

	// initialise the first guess (it's always the same)
	guess := primitives.Word{'t', 'a', 'r', 'e', 's'}

	for {
		// get pattern from guess
		lastPattern := s.Game.Guess(guess)

		// always check for the win before doing anything else
		if s.Game.IsWon() {
			s.Result = guess
			break
		}

		// refine possible answers based on pattern
		remainingAnswers = primitives.Dictionary{}
		for _, word := range s.possibleAnswers {
			pattern := guess.CheckGuess(word)
			if pattern == lastPattern && word != guess {
				remainingAnswers = append(remainingAnswers, word)
			}
		}

		if len(remainingAnswers) == 1 {
			// There's only one choice, we win!
			s.Result = remainingAnswers[0]
			s.Game.Guess(s.Result)
			break
		} else if s.Game.IsLost() {
			// We ran out of guesses.
			break
		} else if len(remainingAnswers) <= 0 {
			// We eliminated all possible answers. This implies some of our guesses
			// were not possible words given the game state at the point they were
			// made.
			log.Fatal(s.allCandidatesEliminated())
		}

		s.possibleAnswers = remainingAnswers

		// go through the guesses and compose them with a score and store in scoredGuesses
		scoredGuesses := make([]guessScore, len(s.possibleAnswers))
		for i, word := range s.possibleAnswers {
			scoredGuesses[i] = guessScore{word, Entropy(word, s.possibleAnswers)}
		}

		// order the guess scoredGuesses by descending score order i.e. best is scoredGuesses[0]
		descendingScore := func(i, j int) bool {
			return scoredGuesses[i].score > scoredGuesses[j].score
		}
		sort.Slice(scoredGuesses, descendingScore)

		// update the guesses
		guess = scoredGuesses[0].guess
	}

	return s.Game
}

func (s *Solver) allCandidatesEliminated() error {
	return fmt.Errorf(
		"--GAME-STATE--: \n%s\n\n--ERROR--:\nEliminated all words;\nBricked it.",
		s.Game,
	)
}

// Entropy calculates the shannon entropy of a guess (A.K.A. the information content). This
// value depends on the number of allowed values. If the dictionary allowed it, the entropy
// (or average information gained with the guess) would be maximised by choosing a guess that
// had p[i] = p, p = 1/N; where N is the number of patterns (3^5 = 81 in our case).
//
//      Note: lim (p*log2(p)) = 0;
//            p->0
//
//      Entropy = Σ  -p[i]log2(p[i])
//               n∈N
//
// Where N is all possible patterns, i.e.
//
//      N = C^5 under the cartesian product; C = {Grey, Yellow, Green}
//
// Note: This quantity is an **Expectation**, not a measure. Once the guess is made and a
// patterns results, the possible correct answers can be narrowed down.
// We can then look at how many times over the pattern halves the number of guess values
// available. Expressed as:
//
//      Information(guess) = -log2(nBefore/nAfter)
//
func Entropy(guess primitives.Word, dict primitives.Dictionary) float64 {
	p := Probabilities(guess, dict)

	bits := 0.0
	for i := range p {
		// This implements the limit of p[i]log2(p[i]) = 0 as p[i] -> 0
		if p[i] == 0.0 {
			continue
		}
		bits += -(p[i] * math.Log2(p[i]))
	}

	return bits
}

// Probabilities calculates a naive probability that the answer will return a given
// pattern by simulating every possible answer and counting the frequency of occurrence
// of each pattern.
func Probabilities(
	guess primitives.Word,
	dict primitives.Dictionary,
) []float64 {
	matches := primitives.Matches(guess, dict)
	distinct := map[primitives.Pattern]int{}
	for _, m := range matches {
		if _, exists := distinct[m.Pattern]; !exists {
			distinct[m.Pattern] = 0
		}
		distinct[m.Pattern]++
	}

	p := make([]float64, util.Pow(3, 5))
	for i := range p {
		frequency, ok := distinct[primitives.PatternFrom(i)]
		if !ok {
			frequency = 0
		}
		p[i] = float64(frequency) / float64(len(dict))
	}
	return p
}
