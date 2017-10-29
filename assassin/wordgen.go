package assassin

import (
	"bufio"
	"io"
	"math/rand"
)

/*
WordGenerator interface for generating player KillWords.
	Gen() generates and returns a single word
*/
type WordGenerator interface {
	Gen() string
}

/*
ListPickWordGenerator is a WordGenerator that selects words from a pre-provided
list of words.
*/
type ListPickWordGenerator struct {
	wordList []string
	current  int
}

// NewListPickWordGenerator creates a ListPickWordGenerator from a given wordList.
func NewListPickWordGenerator(words []string) *ListPickWordGenerator {
	var g = &ListPickWordGenerator{make([]string, 0), 0}
	for _, i := range rand.Perm(len(words)) {
		g.wordList = append(g.wordList, words[i])
	}
	return g
}

/*
ListPickWordGeneratorFromReader creates a ListPickWordGenerator, reading the
contents of the provided r to construct the word list.
*/
func ListPickWordGeneratorFromReader(r io.Reader) (*ListPickWordGenerator, error) {
	var (
		s = bufio.NewScanner(r)
		w = make([]string, 0)
	)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		w = append(w, s.Text())
	}
	return NewListPickWordGenerator(w), s.Err()
}

// Gen picks the next word from the list.
func (g *ListPickWordGenerator) Gen() string {
	var w = g.wordList[g.current]
	g.current = (g.current + 1) % len(g.wordList)
	return w
}
