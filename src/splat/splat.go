package splat

import (
	"bit-wordy/src/primitives"
	"io/ioutil"
	"log"
	"os"
)

const wc = 12972

type Splat []byte

func ReadSplat() *Splat {
	if splat != nil {
		return splat
	}
	file, err := os.Open("data/splat")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	*splat, err = ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return splat
}

var splat *Splat

func (s *Splat) Contract(i int) {

}

func (s *Splat) Pattern(i, j int) primitives.Pattern {
	return primitives.PatternFrom[byte]((*s)[i+wc*j])
}


