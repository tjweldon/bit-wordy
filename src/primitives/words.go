package primitives

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Fivegram is a Word with five letters
type Fivegram [5]byte

// Contains returns true if the Fivegram Contains the character
func (f Fivegram) Contains(character byte) bool {
	for _, letter := range f {
		if character == letter {
			return true
		}
	}
	return false
}

// Matches returns the Pattern when a guess is compared to any other Fivegram
func (f Fivegram) Matches(other Fivegram) Pattern {
	p := DefPattern
	for i := range p {
		if f.Contains(other[i]) {
			p[i] = Yellow
		}
		if other[i] == f[i] {
			p[i] = Green
		}
	}

	return p
}

// Dictionary is a collection of Fivegrams
type Dictionary []Fivegram

// LoadWords pulls the content of the words file into memory
func LoadWords() (dict Dictionary) {
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
		if len(line) != 5 {
			continue
		}
		var word Fivegram
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

// Result is the word and its corresponding pattern as compared to some input
type Result struct {
	Word    Fivegram
	Pattern Pattern
}

func (r Result) String() string {
	s := ""
	for i, color := range r.Pattern {
		switch color {
		case Green:
			s += Green.Paint(r.Word[i])
		case Yellow:
			s += Yellow.Paint(r.Word[i])
		default:
			s += Grey.Paint(r.Word[i])
		}
	}
	return fmt.Sprintf("%5s", s)
}

// ResultSet is a bit pointless
type ResultSet []Result
