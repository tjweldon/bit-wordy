package main

import (
	"bit-wordy/src/cached"
	"bit-wordy/src/games"
	"bit-wordy/src/primitives"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/go-p5/p5"
	"image/color"
	"log"
	"time"
)

const Cache = "data/cache"

var words = primitives.LoadWords()

// chooseAnswer is a lazy pseudo-random answer chooser implementation
// that turns out to be good enough for our purposes.
func chooseAnswer() primitives.Word {
	return words[int(time.Now().UnixMilli())%len(words)]
}

type pair[T any] [2]T

func (p pair[T]) unpack() (T, T) {
	return p[0], p[1]
}

// Guess is an interactive debugging subcommand whose primary usecase is to check the validity
// of the pre-built cache against freshly compared words
type Guess struct {
	Ans   string `arg:"positional"`
	Guess string `arg:"positional"`
}

// Run is the implementation of Guess
func (g *Guess) Run(p *cached.Patterns) {
	guess := primitives.MakeWord(g.Guess)
	ans := primitives.MakeWord(g.Ans)
	results := checkGuess(ans, guess)
	fmt.Printf("FRESH: %s ∙ %s\n", results[1], results[0])
	if p != nil {
		results = fromCache(p, ans, guess)
		fmt.Printf("CACHE: %s ∙ %s\n", results[1], results[0])
	}
}

// Iterate is the subcommand that allows the user to supply a number of games to be solved
// the --print option controls whether we print each game outcome to stdout.
type Iterate struct {
	Times int  `arg:"positional"`
	Print bool `arg:"-p, --print"`
}

// Run is the implementation of Iter
func (i Iterate) Run(p *cached.Patterns) (err error) {
	if !args.Load {
		p, err = cached.LoadPatterns(words)
		if err != nil {
			return err
		}
	}

	var (
		game   = games.NewGame(chooseAnswer())
		answer primitives.Word
		solver = cached.NewSolver(p)
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

type Axis struct {
	Scale      float64
	Start, End pair[float64]
}

func AxisX(start pair[float64], length, scale float64) Axis {
	end := pair[float64]{start[0] + length, start[1]}
	return Axis{
		Scale: scale,
		Start: start,
		End:   end,
	}
}

func AxisY(start pair[float64], length, scale float64) Axis {
	end := pair[float64]{start[0], start[1] - length}
	return Axis{
		Scale: scale,
		Start: start,
		End:   end,
	}
}

func (a Axis) Draw() {
	p5.StrokeWidth(2)
	p5.Stroke(color.Black)
	x1, y1 := a.Start.unpack()
	x2, y2 := a.End.unpack()
	p5.Line(x1, y1, x2, y2)
}

type Renderer struct {
	Angle        float64
	W, H         int
	Margin       float64
	XAxis, YAxis Axis
	Hist         []int
}

func (r *Renderer) Setup() {
	p5.Canvas(r.W, r.H)
	p5.Background(color.Gray{Y: 220})
}

func (r *Renderer) Draw() {
	r.drawAxes()
	i := time.Now().UnixMilli() % 10
	r.Hist[i]++
	r.drawHist()
}

func (r *Renderer) drawAxes() {
	p5.StrokeWidth(2)
	p5.Stroke(color.Black)
	r.XAxis.Draw()
	r.YAxis.Draw()
}

func (r Renderer) drawHist() {
	btm := float64(r.H) - r.Margin
	left := r.Margin
	max := 0.0
	for _, n := range r.Hist {
		if float64(n) >= max {
			max = float64(n)
		}
	}

	p5.StrokeWidth(0)
	for i, n := range r.Hist {
		p5.Fill(color.RGBA{R: 0, G: 175, B: 50, A: 255})
		barHeight := float64(r.H) - 2*r.Margin
		p5.Rect(left+float64(i)*100.0, btm, 99.0, -(barHeight*float64(n))/max)
	}
}

type Visual struct {
	Times int `arg:"positional"`
}

func (v Visual) Run() {
	r := &Renderer{
		W: 1800, H: 600, Margin: 40, Hist: make([]int, 10),
	}
	r.XAxis = AxisX(pair[float64]{r.Margin, float64(r.H) - r.Margin}, float64(r.W)-2.0*r.Margin, 1)
	r.YAxis = AxisY(pair[float64]{r.Margin, float64(r.H) - r.Margin}, float64(r.H)-2.0*r.Margin, 1)
	p5.Run(r.Setup, r.Draw)
	fmt.Println("VISUAL")
}

var args struct {
	Build  bool     `arg:"-b,--build"`
	Dump   bool     `arg:"-d,--dump"`
	Load   bool     `arg:"-l,--load"`
	Play   bool     `arg:"-p,--play"`
	Guess  *Guess   `arg:"subcommand:guess"`
	Iter   *Iterate `arg:"subcommand:iter"`
	Visual *Visual  `arg:"subcommand:visual"`
}

func main() {
	var (
		p   *cached.Patterns
		err error
	)
	arg.MustParse(&args)

	if args.Build {
		log.Println("Building...")
		p = cached.BuildPatterns(words)
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
		p, err = cached.LoadPatterns(words)
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
	if args.Visual != nil {
		args.Visual.Run()
	}

	fmt.Println("Done!")
}

func solveOne(answer primitives.Word, p *cached.Patterns) (*games.Game, *cached.FastSolver, time.Duration) {
	g := games.NewGame(answer)
	s := cached.NewSolver(p)
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

func fromCache(p *cached.Patterns, ans, guess primitives.Word) []primitives.Result {
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
