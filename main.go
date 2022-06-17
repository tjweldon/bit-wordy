package main

import (
	"bit-wordy/src/histogram"
	"bit-wordy/src/primitives"
	"bytes"
	"fmt"
	"github.com/manifoldco/promptui"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func loadWords() (dict patterns.Dictionary) {
	file, err := os.Open("./data/words")
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	lines := bytes.Split(content, []byte{'\n'})

	for _, line := range lines {
		var word patterns.Fivegram
		for i := range word {
			word[i] = line[i]
		}

		dict = append(
			dict,
			word,
		)
	}

	return dict
}

var patternSum = func(rowA, rowB *patterns.Result) bool {
	return rowA.Pattern.Sum() < rowB.Pattern.Sum()
}

var words = loadWords()

func main() {
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
		fmt.Printf("%v\n", err)
		return
	}
	guess := patterns.Fivegram{}
	for i := range guess {
		guess[i] = guessStr[i]
	}

	asSlice := patterns.Matches(guess, words)

	desc := func(b patterns.By) patterns.By {
		return func(p1, p2 *patterns.Result) bool {
			return !b(p1, p2)
		}
	}

	desc(patternSum).Sort(asSlice)

	hist := buildHistogram(asSlice)

	fmt.Println(hist)
}

func buildHistogram(asSlice patterns.ResultSet) histogram.Histogram {
	lastP := asSlice[0].Pattern
	hist := histogram.Histogram{}
	bar := histogram.Bar{}
	for _, res := range asSlice {
		// populate histogram
		if res.Pattern != lastP {
			hist[lastP] = append([]patterns.Result{}, bar...)
			bar, lastP = []patterns.Result{}, res.Pattern
		} else {
			bar = append(bar, res)
		}
	}
	return hist
}
