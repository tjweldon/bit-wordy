package main

import (
	"bit-wordy/src/games"
	"bit-wordy/src/primitives"
	"bit-wordy/src/solver"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

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

func chooseAnswer() primitives.Fivegram {
	return words[int(time.Now().UnixNano())%len(words)]
}

func main() {

}

func DoSplat() {
	out, err := os.OpenFile("data/splat", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o744)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	for _, word := range words {
		for _, werd := range words {
			writer.WriteByte(byte(word.Matches(werd).Sum()))
		}
		writer.Flush()
	}
}

func solveBatch(iterations int) {

	hist := map[int]int{}
	for i := 0; i < 20; i++ {
		hist[i] = 0
	}

	solvers := make([]*solver.Solver, iterations)
	for i := range solvers {
		solvers[i] = solver.NewSolver(games.NewGame(chooseAnswer()), words)
	}

	for j := range solvers {
		solvers[j].Solve()
		hist[solvers[j].GameScore()]++
		fmt.Printf("%-2d%%\r", (100*j)/iterations)
	}

	fmt.Println("Game score distribution")
	for gameScore := 0; gameScore < 10; gameScore++ {
		times, hasOccurred := hist[gameScore]
		if !hasOccurred {
			times = 0
		}

		fmt.Printf("%2d |%s %3d\n", gameScore, strings.Repeat("█", times/2), times)
	}
}

func solveOne() {
	s := solver.NewSolver(games.NewGame(chooseAnswer()), words)
	s.Solve()
}
