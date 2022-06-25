package cached

import (
	"bit-wordy/src/games"
	"bit-wordy/src/primitives"
	"fmt"
	"math"
	"time"
)

type GuessOutcome struct {
	res                      primitives.Result
	expectedInfo, actualInfo float64
	before, after            int
}

func MakeOutcome(score float64, result primitives.Result, prev, current *Patterns) GuessOutcome {
	before, after := len(prev.Vocab), len(current.Vocab)
	return GuessOutcome{
		res:          result,
		expectedInfo: score,
		actualInfo:   -math.Log2(float64(after) / float64(before)),
		before:       before,
		after:        after,
	}
}

func (gO GuessOutcome) String() string {
	return fmt.Sprintf(
		"%s:\n\tE(I): %.2f \t-> I: %.2f\n\tN(ans): %d \t-> N(ans): %d\n",
		gO.res,
		gO.expectedInfo, math.Abs(gO.actualInfo),
		gO.before, gO.after,
	)
}

type FastSolver struct {
	Initial       *Patterns
	prev          *Patterns
	current       *Patterns
	guessMetadata []GuessOutcome
}

func NewSolver(initial *Patterns) *FastSolver {
	return &FastSolver{Initial: initial, current: initial}
}

func (f *FastSolver) Reset() {
	f.current = f.Initial
	f.prev = nil
	f.guessMetadata = []GuessOutcome{}
}

func (f *FastSolver) Solve(g *games.Game) (*games.Game, time.Duration) {
	start := time.Now()
	bestGuess, topScore := primitives.MakeWord("tares"), 0.0
	f.guessOne(g, bestGuess, topScore)
	for !(g.IsWon() || g.IsLost()) {
		bestGuess, topScore := f.current.GetBestGuess()
		f.guessOne(g, bestGuess, topScore)
	}
	playDuration := time.Now().Sub(start)
	return g, playDuration
}

func (f *FastSolver) guessOne(g *games.Game, bestGuess primitives.Word, topScore float64) {
	result := primitives.Result{Pattern: g.Guess(bestGuess), Word: bestGuess}
	f.current, f.prev = f.current.PruneAnswers(result), f.current
	outcome := MakeOutcome(topScore, result, f.prev, f.current)
	f.guessMetadata = append(f.guessMetadata, outcome)
}

func (f FastSolver) String() string {
	s := ""
	for _, oc := range f.guessMetadata {
		s += oc.String()
	}

	return s
}
