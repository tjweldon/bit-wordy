package splat

import (
	"bit-wordy/src/primitives"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
)

type Splat struct {
	Patterns  []byte
	wordcount int
	IndexMap  []int
}

func ReadSplat(length int) *Splat {
	file, err := os.Open("data/splat")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	patterns, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	indexMap := make([]int, length)
	for i := range indexMap {
		indexMap[i] = i
	}

	return &Splat{Patterns: patterns, wordcount: length, IndexMap: indexMap}
}

var splat *Splat

func (s *Splat) RefineBy(guess int, pattern byte) *Splat {
	indices := []int{}
	for i, p := range s.PatternsFor(guess) {
		if p == pattern {
			indices = append(indices, s.IndexMap[i])
		}
	}

	refinedPatterns := make([]byte, len(indices)*len(indices))
	for a, guessIdx := range indices {
		for g, ansIdx := range indices {
			refinedPatterns[a+len(indices)*g] = s.PatternByte(ansIdx, guessIdx)
		}
	}

	return &Splat{refinedPatterns, len(indices), indices}
}

func (s *Splat) PatternsFor(guess int) []byte {
	patternsForWord := make([]byte, s.wordcount)
	for ansIdx := range patternsForWord {
		patternsForWord[ansIdx] = s.PatternByte(ansIdx, guess)
	}
	return patternsForWord
}

func (s Splat) WordCount() int {
	return s.wordcount
}

func (s *Splat) Frequencies(wordIndex int) [primitives.PatternCardinality]int {
	freqs := [primitives.PatternCardinality]int{}
	for l := 0; l < s.wordcount; l++ {
		freqs[s.PatternByte(l, wordIndex)]++
	}

	return freqs
}

func (s *Splat) Probabilities(wordIndex int) [primitives.PatternCardinality]float64 {
	freqs := [primitives.PatternCardinality]float64{}
	occurrence := 1.0 / float64(s.wordcount)
	for ansIdx := 0; ansIdx < s.wordcount; ansIdx++ {
		freqs[s.Patterns[ansIdx+s.wordcount*wordIndex]] += occurrence
	}

	return freqs
}

func (s *Splat) Entropy(wordIndex int) float64 {
	acc := 0.0
	for _, p := range s.Probabilities(wordIndex) {
		acc += -(p * math.Log2(p))
	}

	return acc
}

func (s *Splat) Pattern(ansIdx, guessIdx int) primitives.Pattern {
	return primitives.PatternFrom[byte](s.PatternByte(ansIdx, guessIdx))
}

func (s *Splat) String() string {
	st := ""
	for i := range s.IndexMap {
		for j := range s.IndexMap {
			st += fmt.Sprintf("%s ", primitives.PatternFrom[byte](s.PatternByte(i, j)))
		}
		st += "\n"
	}

	return st
}

func (s *Splat) PatternByte(ansIdx, guessIdx int) byte {
	return s.Patterns[ansIdx+(s.WordCount())*guessIdx]
}
