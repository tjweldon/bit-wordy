package primitives

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Word is a Word with five letters
type Word [5]byte

// MakeWord uses the first five bytes of the input string to populate a fivegram,
// if there are less than five, the remaining bytes are null
func MakeWord(s string) Word {
	res := Word{}
	input := []byte(s)
	if len(input) < 5 {
		input = append(input, make([]byte, 5-len(input))...)
	}
	for i, b := range input {
		res[i] = b
	}

	return res
}

// Contains returns true if the Word Contains the character
func (f Word) Contains(character byte) bool {
	for _, letter := range f {
		if character == letter {
			return true
		}
	}
	return false
}

// CheckGuess returns the Pattern when a guess is compared to any other Word
func (f Word) CheckGuess(guess Word) Pattern {
	p := DefPattern
	for i := range p {
		if f.Contains(guess[i]) {
			p[i] = Yellow
		}
		if guess[i] == f[i] {
			p[i] = Green
		}
	}

	return p
}

// Dictionary is a collection of Fivegrams
type Dictionary []Word

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
		var word Word
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

// IndexOf returns the index of a Word in the Dictionary
func (d Dictionary) IndexOf(word Word) (idx int, ok bool) {
	for i, w := range d {
		if w == word {
			return i, true
		}
	}

	return -1, false
}

// Result is the word and its corresponding pattern as compared to some input
type Result struct {
	Word    Word
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
