package assassin

import (
	"math/rand"
	"strings"
	"testing"
)

func TestWordList(t *testing.T) {
	var (
		wl  = strings.NewReader("three\n   one   two")
		exp = []string{"one", "two", "three", "one"}
	)
	rand.Seed(0)
	var g, err = WordListFromReader(wl)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		if w := g.Next(); w != exp[i] {
			t.Error("Unexpected word for iteration", i, w, "!=", exp[i])
		}
	}
}
