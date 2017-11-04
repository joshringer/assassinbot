package assassin

import (
	"bufio"
	"io"
	"math/rand"
)

/*
WordGenerator interface for generating player KillWords.
	Next() generates and returns a single word
*/
type WordGenerator interface {
	Next() string
}

/*
WordList is a WordGenerator that selects words from a pre-provided
list of words.
*/
type WordList struct {
	words   []string
	current int
}

// NewWordList creates a WordList from a given words.
func NewWordList(words []string) *WordList {
	var g = &WordList{make([]string, 0), 0}
	for _, i := range rand.Perm(len(words)) {
		g.words = append(g.words, words[i])
	}
	return g
}

/*
WordListFromReader creates a WordList, reading the
contents of the provided r to construct the word list.
*/
func WordListFromReader(r io.Reader) (*WordList, error) {
	var (
		s = bufio.NewScanner(r)
		w = make([]string, 0)
	)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		w = append(w, s.Text())
	}
	return NewWordList(w), s.Err()
}

// Next picks the next word from the list.
func (g *WordList) Next() string {
	var w = g.words[g.current]
	g.current = (g.current + 1) % len(g.words)
	return w
}
