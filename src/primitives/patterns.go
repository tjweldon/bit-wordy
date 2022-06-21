package primitives

import "github.com/fatih/color"

// Color is the actual ansi printed color
type Color color.Attribute

// Paint is a convenience wrapper for coloring strings
func (c Color) Paint(b byte) string {
	s := string([]byte{b})
	return color.New(color.Attribute(c), color.FgBlack, color.Faint).SprintfFunc()("%s", s)
}

// the colors mean:
const (
	// Grey - the letter doesn't appear
	Grey = Color(color.BgWhite)
	// Yellow - the letter is in the Word
	Yellow = Color(color.BgYellow)
	// Green - the letter is in the position
	Green = Color(color.BgGreen)
)

const PatternCardinality = 243

// Pattern is the letter Result
type Pattern [5]Color

func PatternFrom[T int | byte](i T) Pattern {
	p := DefPattern
	j := 0
	cols := []Color{Grey, Yellow, Green}
	for i > 0 {
		digit := i % 3
		i = (i - digit) / 3
		p[j] = cols[digit]
		j++
	}

	return p
}

func (p Pattern) Sum() (s int8) {
	base := [5]int8{1, 3, 9, 27, 81}
	for i, color := range p {
		var digit int8
		switch color {
		case Green:
			digit = 2
		case Yellow:
			digit = 1
		default:
			digit = 0
		}

		s += digit * base[i]
	}

	return s
}

func (p Pattern) Byte() byte {
	return byte(p.Sum())
}

func (p Pattern) String() string {
	return Result{Word: FromStr("#####"), Pattern: p}.String()
}

var (
	DefPattern = Pattern{Grey, Grey, Grey, Grey, Grey}
	Win        = Pattern{Green, Green, Green, Green, Green}
)

// Matches computes the pattern for each word in the dictionary and returns them
func Matches(guess Fivegram, dict Dictionary) ResultSet {
	results := ResultSet{}
	for _, word := range dict {
		results = append(
			results,
			Result{
				Word:    word,
				Pattern: word.CheckGuess(guess),
			},
		)
	}
	return results
}
