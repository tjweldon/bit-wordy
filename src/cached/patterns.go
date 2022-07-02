package cached

import (
	"bit-wordy/src/primitives"
	"fmt"
	"github.com/kelindar/binary"
	"io/ioutil"
	"log"
	"math"
	"os"
)

const Cache = "data/cache"

type Patterns struct {
	Vocab        primitives.Dictionary
	index        *map[primitives.Word]int
	patternCache [][]byte
	patternIndex *primitives.PatternSpace
	fastLog      *FastLog
}

// BuildPatterns is the computation of all comparisons and the storage of the results
func BuildPatterns(dict primitives.Dictionary) *Patterns {
	patternCache := make([][]byte, len(dict))
	// index := map[primitives.Word]int{}
	for guessId, guess := range dict {
		// index[answer] = guessId

		guessOutcomes := make([]byte, len(dict))
		for ansId, answer := range dict {
			guessOutcomes[ansId] = answer.CheckGuess(guess).Byte()
		}
		patternCache[guessId] = guessOutcomes
	}

	p := &Patterns{
		Vocab:        dict,
		patternCache: patternCache,
	}
	p.PopulateIndices(dict)
	return p
}

func (p *Patterns) PopulateIndices(vocab primitives.Dictionary) {
	p.Vocab = vocab

	p.index = &map[primitives.Word]int{}
	for i, answer := range p.Vocab {
		(*p.index)[answer] = i
	}
}

// LoadPatterns will build the patterns from the cached computed value if the cache exists
func LoadPatterns(dict primitives.Dictionary) (*Patterns, error) {
	file, err := os.Open(Cache)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	fresh := Patterns{}

	c := make([][]byte, len(dict))
	err = binary.Unmarshal(content, &c)
	if err != nil {
		return nil, err
	}

	fresh.patternCache = c
	fresh.fastLog = NewFastLog(dict)
	fresh.PopulateIndices(dict)

	return &fresh, nil
}

// Dump does binary serialisation to file
func (p *Patterns) Dump(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o744)
	if err != nil {
		return err
	}
	defer file.Close()

	err = binary.MarshalTo(p.patternCache, file)
	if err != nil {
		return err
	}

	return nil
}

// Compare is the basic primitives.Word × primitives.Word -> primitives.Pattern mapping
// NOTE: this is not symmetric e.g.
//
//      Compare(obese, emacs) |-> Pattern(00111),
//      Compare(emacs, obese) |-> Pattern(10001);
//
//      Grey=0,
//      Yellow=1;
//
func (p *Patterns) Compare(guess, ans primitives.Word) primitives.Pattern {
	iGuess, iAns := (*p.index)[guess], (*p.index)[ans]
	return primitives.PatternFrom(p.patternCache[iGuess][iAns])
}

type FastLog struct {
	cache []float64
}

func NewFastLog(dict primitives.Dictionary) *FastLog {
	return &FastLog{make([]float64, len(dict)+1)}
}

func (f *FastLog) Log2(x int) float64 {
	if x == 1 {
		return 0.0
	}
	if f.cache[x] == 0.0 {
		f.cache[x] = math.Log2(float64(x))
	}
	return f.cache[x]
}

// Entropies is the slice of estimated information gained in bits by a guess, indexed by guess id.
// Information content defined as follows:
//
//      After: Remaining guesses After receiving the pattern and pruning eliminated answers from B.
//      Before: The possible answer candidates Before the guess.
//
// Then the **actual** information of the guess after it is made is:
//
//      Information(guess) = -log2( len(After) / len(Before) )
//
// This can be interpreted as the number of times over the solution space's size was halved by the guess.
// If len(Before) = 8 and len(After) = 1, and h(n) = (1/2) * n, then
//
//      n = h(n);   n: 8 -> 4
//      n = h(n);   n: 4 -> 2
//      n = h(n);   n: 2 -> 1
//
// Gives 3 times over the guess halved the number of possible answers, or 3 'bits' of information.
//
// Each value is the shannon entropy of the system when the corresponding guess when the answer is
// provided, and the pattern is not yet known. A large value corresponds to a more even distribution
// of outcomes and is maximised by
//
//      p[i] = p[j] = 1/len(guesses)     for     i,j ∈ [: len(guesses)]
//
// TL;DR: For the purposes of solving wordle, we are looking to maximise the change of entropy on receiving
// the pattern and pruning answers. That amounts to choosing the guess with the greatest
func (p Patterns) Entropies() []float64 {
	wordCount := len(p.Vocab)

	// for each guess, create the pattern frequency distribution
	frequencies := make([][]int, wordCount)
	for guessId, answers := range p.patternCache {
		patterns := make([]int, len(p.patternIndex))
		frequencies[guessId] = patterns
		for _, pattern := range answers {
			patterns[pattern]++
		}
	}

	// for each guess, estimate the p(guess*ans=pattern)=Count(pattern)/Count(anwer)
	entropies := make([]float64, wordCount)
	for guessId, patternFreqs := range frequencies {
		entropy := 0.0
		for _, patternCount := range patternFreqs {
			// this is the implementation lim p*log2(p) = 0 as p->0, and avoids
			// pointless computation where the answer will be NaN.
			if patternCount == 0 {
				continue
			}

			// this is P(pattern) = occurrences(pattern) / count(guesses)
			probability := float64(patternCount) / float64(wordCount)
			// entropy += -(p * math.Log2(p))
			// here we add the contribution of this outcome to the total for the guess
			entropy += -(probability * (p.fastLog.Log2(patternCount) - p.fastLog.Log2(wordCount)))
		}
		entropies[guessId] = entropy
	}

	return entropies
}

// GetBestGuess returns the highest scoring guess, along with its score
func (p Patterns) GetBestGuess() (bestGuess primitives.Word, topScore float64) {
	bestGuessId := 0
	for guessId, score := range p.Entropies() {
		if score > topScore {
			bestGuessId, topScore = guessId, score
		}
	}

	bestGuess = p.Vocab[bestGuessId]
	return bestGuess, topScore
}

// PruneAnswers returns
func (p *Patterns) PruneAnswers(result primitives.Result) *Patterns {
	newVocab := primitives.Dictionary{}
	patternByte := result.Pattern.Byte()
	guessId := (*p.index)[result.Word]
	patternBytes := p.patternCache[guessId]
	idMap := []int{}
	for ansId, pattern := range patternBytes {
		if pattern == patternByte {
			newVocab = append(newVocab, p.Vocab[ansId])
			idMap = append(idMap, ansId)
		}
	}

	if len(newVocab) == 0 {
		log.Fatal(fmt.Errorf("result %s gave no remaining answers after prune, from wordcount of %d", result, len(p.Vocab)))
	}

	// deriving the new pattern cache from the previous game state here is 3x faster
	// than supplying the vocabulary and repeating the checks for the refined set
	newCache := make([][]byte, len(newVocab))
	for newGuessId, oldGuessId := range idMap {
		ansRow := make([]byte, len(idMap))
		for newAnsId, oldAnsId := range idMap {
			ansRow[newAnsId] = p.patternCache[oldGuessId][oldAnsId]
		}
		newCache[newGuessId] = ansRow
	}
	newIndex := &map[primitives.Word]int{}
	for i, answer := range newVocab {
		(*newIndex)[answer] = i
	}

	patterns := &Patterns{
		Vocab: newVocab, index: newIndex, patternCache: newCache, patternIndex: p.patternIndex, fastLog: p.fastLog,
	}

	// patterns := BuildPatterns(newVocab)
	return patterns
}
