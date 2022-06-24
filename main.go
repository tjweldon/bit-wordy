package main

import (
	"bit-wordy/src/games"
	"bit-wordy/src/primitives"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/kelindar/binary"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

const Cache = "data/cache"

var words = primitives.LoadWords()

func chooseAnswer() primitives.Word {
	return words[int(time.Now().UnixMilli())%len(words)]
}

type pair[T any] [2]T

func (p pair[T]) unpack() (T, T) {
	return p[0], p[1]
}

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

var fast = NewFastLog(words)

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

	// for each guess, estimate the p(guess*ans=pattern)=Count(pattern)/Count(anwer
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
			p := float64(patternCount) / float64(wordCount)
			entropy += -(p * (fast.Log2(patternCount) - fast.Log2(wordCount)))
			// here we add the contribution of this outcome to the total for the guess
			// entropy += -(p * math.Log2(p))
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

	patterns := &Patterns{Vocab: newVocab, index: newIndex, patternCache: newCache, patternIndex: p.patternIndex}

	// patterns := BuildPatterns(newVocab)
	return patterns
}

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

type Guess struct {
	Ans   string `arg:"positional"`
	Guess string `arg:"positional"`
}

func (g *Guess) Run(p *Patterns) {
	guess := primitives.MakeWord(g.Guess)
	ans := primitives.MakeWord(g.Ans)
	results := checkGuess(ans, guess)
	fmt.Printf("FRESH: %s ∙ %s\n", results[1], results[0])
	if p != nil {
		results = fromCache(p, ans, guess)
		fmt.Printf("CACHE: %s ∙ %s\n", results[1], results[0])
	}
}

type Iterate struct {
	Times int  `arg:"positional"`
	Print bool `arg:"-p, --print"`
}

func (i Iterate) Run(p *Patterns) (err error) {
	if !args.Load {
		p, err = LoadPatterns(words)
		if err != nil {
			return err
		}
	}

	var (
		game   = games.NewGame(chooseAnswer())
		answer primitives.Word
		solver = NewSolver(p)
	)
	prin := func(d time.Duration, g *games.Game) {}
	if i.Print {
		prin = func(d time.Duration, g *games.Game) {
			fmt.Printf("TIME: %s\n", d.String())
			fmt.Printf("GAME:\n%s\n", g)
		}

	}
	start := time.Now()
	for j := 0; j < i.Times; j++ {
		solver.Reset()
		answer = chooseAnswer()
		game.Reset(answer)
		game = games.NewGame(answer)
		_, playDuration := solver.Solve(game)
		prin(playDuration, game)
	}
	fmt.Println(time.Now().Sub(start) / time.Duration(i.Times))
	return err
}

var args struct {
	Build bool     `arg:"-b,--build"`
	Dump  bool     `arg:"-d,--dump"`
	Load  bool     `arg:"-l,--load"`
	Play  bool     `arg:"-p,--play"`
	Guess *Guess   `arg:"subcommand:guess"`
	Iter  *Iterate `arg:"subcommand:iter"`
}

func main() {
	var (
		p   *Patterns
		err error
	)
	arg.MustParse(&args)

	if args.Build {
		log.Println("Building...")
		p = BuildPatterns(words)
		log.Println("Built!")
	}
	if args.Dump {
		log.Println("Dumping...")
		if err = p.Dump(Cache); err != nil {
			log.Fatal(err)
		}
		log.Println("Dumped!")
	}
	if args.Load {
		log.Println("Loading...")
		p, err = LoadPatterns(words)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Loaded!")
	}
	if args.Play {
		log.Println("Playing...")

		answer := chooseAnswer()
		fmt.Printf("Chose answer: %s\n", answer)
		g, s, playDuration := solveOne(answer, p)

		fmt.Printf("%s\n", playDuration.String())
		fmt.Printf("GAME:\n%s\n", g)
		fmt.Printf("SOLVER:\n%s", s)

		log.Println("Played!")
	}
	if args.Guess != nil {
		args.Guess.Run(p)
	}
	if args.Iter != nil {
		err = args.Iter.Run(p)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Done!")
}

func solveOne(answer primitives.Word, p *Patterns) (*games.Game, *FastSolver, time.Duration) {
	g := games.NewGame(answer)
	s := NewSolver(p)
	g, playDuration := s.Solve(g)
	s.Reset()
	return g, s, playDuration
}

func checkGuess(ans primitives.Word, guess primitives.Word) []primitives.Result {
	results := []primitives.Result{
		{
			Pattern: ans.CheckGuess(guess),
			Word:    guess,
		},
		{
			Pattern: ans.CheckGuess(guess),
			Word:    ans,
		},
	}
	return results
}

func fromCache(p *Patterns, ans, guess primitives.Word) []primitives.Result {
	results := []primitives.Result{
		{
			Pattern: p.Compare(guess, ans),
			Word:    guess,
		},
		{
			Pattern: p.Compare(guess, ans),
			Word:    ans,
		},
	}
	return results
}
