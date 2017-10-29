package assassin

import (
	"strings"
	"testing"
)

type dummyWordGen struct {
	b byte
}

func (d *dummyWordGen) Gen() string {
	var c = string(d.b)
	d.b++
	return c + c + c + c
}

func TestListPickWordGen(t *testing.T) {
	var (
		wl  = strings.NewReader("two\n   one   three")
		exp = []string{"one", "two", "three", "one"}
	)
	var g, err = ListPickWordGeneratorFromReader(wl)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		if w := g.Gen(); w != exp[i] {
			t.Error("Unexpected Gen order for iteration", i, w, "!=", exp[i])
		}
	}
}
