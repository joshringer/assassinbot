package assassin

import "testing"

func TestPlayer(t *testing.T) {
	var (
		wg      = &WordList{words: []string{"aaaa", "bbbb", "cccc", "dddd", "eeee"}}
		a, b, c *Player
	)

	// Test init
	t.Run("NewPlayer", func(t *testing.T) {
		a = NewPlayer(1, "A", wg)
		b = NewPlayer(2, "B", wg)
		c = NewPlayer(3, "C", wg)

		if a.ID != 1 || a.Name != "A" || !a.Alive {
			t.Error(a, "not initialised as expected")
		}
	})

	// Test target set
	t.Run("SetTarget", func(t *testing.T) {
		a.SetTarget(b)
		b.SetTarget(c)
		c.SetTarget(a)
		if a.target != b {
			t.Error(a, "target not set to", b)
		}
		if b.contract != a {
			t.Error(b, "contract not set to", a)
		}
		if a.KillWord != "aaaa" {
			t.Error(a, "KillWord not set to aaaa")
		}
	})

	// Test eliminate
	t.Run("SetEliminated", func(t *testing.T) {
		b.SetEliminated()
		if a.target != c {
			t.Error(a, "target not updated to", c)
		}
		if c.contract != a {
			t.Error(c, "contract not updated to", a)
		}
		if a.KillWord != "dddd" {
			t.Error(a, "KillWord not updated to dddd")
		}
	})

	// TODO: test this?
	c.SetTarget(nil)

	// Test gets
	t.Run("GetTarget/Contract", func(t *testing.T) {
		var np = Player{}
		if at, ok := a.GetTarget(); at != *c || !ok {
			t.Error(a, "unexpected GetTarget response", at, ok)
		}
		if ct, ok := c.GetTarget(); ct != np || ok {
			t.Error(c, "non-empty GetTarget response", ct, ok)
		}
		if cc, ok := c.GetContract(); cc != *a || !ok {
			t.Error(c, "unexpected GetContract response", cc, ok)
		}
		if ac, ok := a.GetContract(); ac != np || ok {
			t.Error(a, "non-empty GetContract response", ac, ok)
		}
	})
}

func TestGame(t *testing.T) {
	var (
		pl = map[ID]string{1: "A", 2: "B", 3: "C"}
		wg = NewWordList([]string{"aaaa", "bbbb", "cccc"})
		g  *Game
	)

	// Test init
	t.Run("NewGame", func(t *testing.T) {
		g = NewGame(1, pl, wg)
		if g.ID != 1 || len(g.players) != 3 {
			t.Error(g, "not initialised as expected")
		}
	})

	// Test gets
	t.Run("GetPlayer", func(t *testing.T) {
		var np = Player{}
		if a, ok := g.GetPlayer(1); a.ID != 1 || !ok {
			t.Error(g, "unexpected GetPlayer response", a, ok)
		}
		if n, ok := g.GetPlayer(123); n != np || ok {
			t.Error(g, "non-empty GetPlayer response", n, ok)
		}
	})
	t.Run("WithPlayers", func(t *testing.T) {
		var (
			c  = 0
			l  = len(g.players)
			np = Player{}
		)
		g.WithPlayers(func(p Player) {
			if p == np {
				t.Error("unexpected empty player")
			}
			c++
		})
		if c != l {
			t.Error("loop iteration", c, "!= player count", l)
		}
	})

	// Test start
	t.Run("Start", func(t *testing.T) {
		g.Start()
		var p, ok = g.players[1]
		if !ok || p == nil {
			t.Error("Player not assigned")
		} else if p.target == nil {
			t.Error("Player missing target")
		} else if p.target.target == nil {
			t.Error("Player target missing target")
		} else if p.target.target.target != p {
			t.Error(p, "->", p.target, "->", p.target.target, "->", p.target.target.target)
		}
	})
}
