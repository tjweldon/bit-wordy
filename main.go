package main

import (
	"bit-wordy/src/games"
	"bit-wordy/src/histogram"
	"bit-wordy/src/primitives"
	"bit-wordy/src/solver"
	"fmt"
	"github.com/manifoldco/promptui"
	"log"
	"strings"
	"time"
)

var patternSum = func(rowA, rowB *primitives.Result) bool {
	return rowA.Pattern.Sum() < rowB.Pattern.Sum()
}

var words = primitives.LoadWords()

var answer = words[int(time.Now().UnixNano())%len(words)]

// PatternFrequency encapsulates the frequency distribution of patterns. It can be interpreted as follows:
//  1. Given a dictionary and a guess, we can compute the patterns by comparing the guess to each word in the dictionary in turn.
//  2. This can be turned into a function f: (word ⊗ dictionary) -> (pattern |-> frequency)
//  3. The result of which is the frequency distribution across patterns for that word
type PatternFrequency map[primitives.Pattern]int

func patternFrequency(guess primitives.Fivegram, dict primitives.Dictionary) PatternFrequency {
	matches := primitives.Matches(guess, dict)
	distinct := PatternFrequency{}
	for _, m := range matches {
		if _, exists := distinct[m.Pattern]; !exists {
			distinct[m.Pattern] = 0
		}
		distinct[m.Pattern]++
	}

	return distinct
}

func (pf PatternFrequency) String() (s string) {
	i := 0
	for patt, freq := range pf {
		c := fmt.Sprintf("%-5d", freq)
		word := primitives.Fivegram{c[0], c[1], c[2], c[3], c[4]}

		separator := "\t"
		if i%10 == 9 {
			separator = "\n"
		}
		s += primitives.Result{
			Word:    word,
			Pattern: patt,
		}.String() + separator
		i++
	}

	return s
}

func sum(all ...float64) (total float64) {
	for _, term := range all {
		total += term
	}
	return total
}

func main() {
	hist := map[int]int{}
	for i := 0; i < 20; i++ {
		hist[i] = 0
	}

	chooseAnswer := func() primitives.Fivegram {
		return words[int(time.Now().UnixNano())%len(words)]
	}

	const iterations = 2000
	solvers := make([]*solver.Solver, iterations)
	for i := range solvers {
		solvers[i] = solver.NewSolver(games.NewGame(chooseAnswer()), words)
	}

	for j := range solvers {
		solvers[j].Solve()
		hist[solvers[j].GameScore()]++
		fmt.Printf("%-2d%%\r", (100*j)/iterations)
		// if j%10 == 0 {
		// }
		// fmt.Println(solvers[j])
		// fmt.Println()
	}

	fmt.Println("Game score distribution")
	for gameScore := 0; gameScore < 10; gameScore++ {
		times, hasOccurred := hist[gameScore]
		if !hasOccurred {
			times = 0
		}

		fmt.Printf("%2d |%s %3d\n", gameScore, strings.Repeat("█", times/2), times)
	}

	// game := games.NewGame(chooseAnswer())
	// s := solver.NewSolver(game, words)
	// s.Solve()
	// fmt.Println(s)

	// Wordle()
}

func Wordle() {
	possibleAnswers := words
	fmt.Printf("%s\n", answer)
	var (
		guess   primitives.Result
		guesses []primitives.Result
	)
	for guess.Word != answer {
		guess = userGuess()
		guesses = append(guesses, guess)
		allPatterns := primitives.Matches(guess.Word, possibleAnswers)
		hist := buildHistogram(allPatterns)
		bar := hist[guess.Pattern]
		newWords := make(primitives.Dictionary, len(bar))
		for i, res := range bar {
			newWords[i] = res.Word
		}
		possibleAnswers = newWords

		for _, g := range guesses {
			fmt.Println(g)
		}
	}
}

func userGuess() primitives.Result {
	guessPrompt := promptui.Prompt{
		Label: "Guess",
		Validate: func(s string) error {
			for _, r := range s {
				if !strings.Contains("abcdefghijklmnopqrstuvwxyz", string(r)) {
					return fmt.Errorf("Guess must match pattern /[a-z]{5}/")
				}
			}
			return nil
		},
	}
	guessStr, err := guessPrompt.Run()

	if err != nil {
		log.Fatalf("%v\n", err)
	}
	guess := primitives.Fivegram{}
	for i := range guess {
		guess[i] = guessStr[i]
	}

	answerResult := primitives.Result{Word: guess, Pattern: answer.Matches(guess)}
	return answerResult
}

func buildHistogram(asSlice primitives.ResultSet) histogram.Histogram {
	hist := histogram.Histogram{}
	for i := 0; i < 3*3*3*3*3; i++ {
		hist[primitives.FromInt(i)] = histogram.Bar{}
	}
	for _, res := range asSlice {
		bar := hist[res.Pattern]
		bar = append(bar, res)
		hist[res.Pattern] = bar
	}
	return hist
}
