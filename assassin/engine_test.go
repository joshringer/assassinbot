package assassin

import (
	"errors"
	"regexp"
	"testing"
	"time"
)

func TestAttackQueue(t *testing.T) {
	var q = newAttackQueue()
	t.Run("push", func(t *testing.T) {
		q.push(1, 2)
		q.push(3, 4)
		if len(q.q) != 2 {
			t.Error("Push to q failed")
		}
	})
	t.Run("each", func(tt *testing.T) {
		var c = 0
		q.each(func(p, t ID, r bool) bool {
			c++
			if r {
				tt.Error("Expected r = false for", p, t)
			}
			return p == 1
		})
		if c != 2 {
			tt.Error("Iterated", c, "times, expected 2")
		}
		q.each(func(p, t ID, r bool) bool {
			if p == 1 && !r || p != 1 && r {
				tt.Error("Expected r =", !r, "for", p, t)
			}
			return r
		})
	})
	t.Run("pop", func(tt *testing.T) {
		var p, t, r = q.pop()
		if p != 1 || t != 2 || !r {
			tt.Error("Expected 1, 2, true, got", p, t, r)
		}
		var p2, t2, r2 = q.pop()
		if p2 != 3 || t2 != 4 || r2 {
			tt.Error("Expected 3, 4, false, got", p2, t2, r2)
		}
	})
}

func timeout(d time.Duration) chan time.Duration {
	var c = make(chan time.Duration)
	go func() { time.Sleep(d); c <- d }()
	return c
}

type testMessageHandler struct {
	t *testing.T
	a chan string
	n chan struct {
		Player
		string
	}
}

func newTestMessageHandler(t *testing.T) *testMessageHandler {
	return &testMessageHandler{t, make(chan string), make(chan struct {
		Player
		string
	})}
}

func (h *testMessageHandler) set(t *testing.T) {
	h.t = t
}

type playerString struct {
	Player
	string
}

type playerRegexp struct {
	Player
	*regexp.Regexp
}

func (h *testMessageHandler) expect(m ...interface{}) {
	var to = timeout(time.Second)
	for len(m) > 0 {
		select {
		case s := <-h.a:
			var u = true
			for i := 0; i < len(m); i++ {
				switch cmp := m[i].(type) {
				case string:
					if s != cmp {
						h.t.Error("Announcement", s, "!=", cmp)
					}
				case *regexp.Regexp:
					if !cmp.MatchString(s) {
						h.t.Error("Announcement", s, "!~", cmp)
					}
				default:
					continue
				}
				m = append(m[:i], m[i+1:]...)
				u = false
				break
			}
			if u {
				h.t.Error("Announcement", s, "unexpected")
			}
		case v := <-h.n:
			var u = true
			for i := 0; i < len(m); i++ {
				switch cmp := m[i].(type) {
				case playerString:
					if v.Player.ID != cmp.Player.ID {
						continue
					}
					if v.string != cmp.string {
						h.t.Error("Notification", v.string, "to", v.Player, "!=", cmp.string)
					}
				case playerRegexp:
					if v.Player.ID != cmp.Player.ID {
						continue
					}
					if !cmp.Regexp.MatchString(v.string) {
						h.t.Error("Notification", v.string, "to", v.Player, "!~", cmp.Regexp)
					}
				default:
					continue
				}
				m = append(m[:i], m[i+1:]...)
				u = false
				break
			}
			if u {
				h.t.Error("Notification", v.string, "to", v.Player, "unexpected")
			}
		case d := <-to:
			h.t.Error("Expected messages", m, "not seen within", d)
			return
		}
	}
}

func (h *testMessageHandler) Announce(s string) {
	h.t.Logf("# >> %v", s)
	var to = timeout(3 * time.Second)
	select {
	case h.a <- s:
	case <-to:
		h.t.Error("Announcement", s, "unexpected")
	}
}

func (h *testMessageHandler) Notify(p Player, s string) {
	h.t.Logf("@%v >> %v", p.Name, s)
	var to = timeout(3 * time.Second)
	select {
	case h.n <- struct {
		Player
		string
	}{p, s}:
	case <-to:
		h.t.Error("Notification", s, "to", p, "unexpected")
	}
}

type triggeredTimingFunc struct {
	t    *testing.T
	wait chan time.Duration
}

func newTriggeredTimingFunc(t *testing.T) *triggeredTimingFunc {
	return &triggeredTimingFunc{t, make(chan time.Duration)}
}

func (ttf *triggeredTimingFunc) set(t *testing.T) {
	ttf.t = t
}

func (ttf *triggeredTimingFunc) Calc() time.Duration {
	var to = timeout(time.Second)
	select {
	case d := <-ttf.wait:
		return d
	case <-to:
		ttf.t.Error("Timeout waiting for trigger")
		return 0
	}

}

func TestGameEngineBasic(t *testing.T) {
	var mh = newTestMessageHandler(t)
	var tf = newTriggeredTimingFunc(t)
	var e = NewGameEngine(LangEn, mh, tf)
	var g = NewGame(1, map[ID]string{1: "A"}, NewWordList([]string{"aaaa"}))
	var res = make(chan error)
	go func() { res <- e.Run(g) }()
	go func() {
		time.Sleep(10 * time.Second)
		res <- errors.New("Game run timeout")
	}()
	mh.expect(LangEn.GS)
	if p, ok := g.GetPlayer(1); ok {
		mh.expect(playerString{p, LangEn.Fmt(LangEn.PT, "A", "aaaa")})
	} else {
		t.Error("Couldn't get player 1")
	}
	mh.expect(LangEn.GE, LangEn.Fmt(LangEn.GW, "A"))
	var r = <-res
	if r != nil {
		t.Fatal(r)
	}
}

func input(t *testing.T, e *GameEngine, p *Player, s string) {
	t.Logf("@%v << %v", p.Name, s)
	e.IncomingTalk(p.ID, s)
}

func TestGameEngineRunthrough(t *testing.T) {
	var mh = newTestMessageHandler(t)
	var tf = newTriggeredTimingFunc(t)
	var e = NewGameEngine(LangEn, mh, tf)
	var g = NewGame(
		1,
		map[ID]string{1: "Ace", 2: "Bee", 3: "Cee", 4: "Dee"},
		NewWordList([]string{"kw1", "kw2", "kw3", "kw4", "kw5", "kw6", "kw7"}),
	)
	var s *Player
	var rpt = regexp.MustCompile(LangEn.Fmt(LangEn.PT, ".+", ".+"))
	var sm = make([]interface{}, 0)
	for _, p := range g.players {
		sm = append(sm, playerRegexp{*p, rpt})
		s = p
	}
	var res = make(chan error)
	go func() { res <- e.Run(g) }()
	go func() {
		time.Sleep(10 * time.Second)
		res <- errors.New("Game run timeout")
	}()
	mh.expect(LangEn.GS)
	mh.expect(sm...)
	t.Run("attack", func(t *testing.T) {
		mh.set(t)
		tf.set(t)
		var t1 = s.target
		if t1 == nil {
			t.Fatal("Player", s, "missing target")
		}
		input(t, e, s, "Text including "+s.KillWord)
		var to = timeout(time.Second)
		select {
		case tf.wait <- time.Nanosecond:
		case d := <-to:
			t.Error("Wait not requested within", d)
		}
		mh.expect(playerString{*s, LangEn.PAS})
		mh.expect(LangEn.Fmt(LangEn.GD, t1.Name))
		mh.expect(playerRegexp{*s, rpt})
	})
	t.Run("counter", func(t *testing.T) {
		mh.set(t)
		tf.set(t)
		var t1 = s.target
		if t1 == nil {
			t.Fatal("Player", s, "missing target")
		}
		var t2 = t1.target
		if t2 == nil {
			t.Fatal("Player", t1, "missing target")
		}
		input(t, e, t1, "Text including "+t1.KillWord)
		input(t, e, t2, "Response including "+t2.KillWord)
		var to = timeout(time.Second)
		select {
		case tf.wait <- time.Nanosecond:
		case d := <-to:
			t.Error("Wait not requested within", d)
		}
		mh.expect(playerString{*t2, LangEn.PCS})
		mh.expect(LangEn.Fmt(LangEn.GD, t1.Name))
		mh.expect(playerRegexp{*s, rpt})
	})
	t.Run("assassinate", func(t *testing.T) {
		mh.set(t)
		var t1 = s.target
		if t1 == nil {
			t.Fatal("Player", s, "missing target")
		}
		input(t, e, t1, "Text including "+s.KillWord)
		mh.expect(LangEn.Fmt(LangEn.GD, t1.Name))
		mh.expect(playerRegexp{*s, rpt})
	})
	mh.set(t)
	mh.expect(LangEn.GE, LangEn.Fmt(LangEn.GW, s.Name))
	var r = <-res
	if r != nil {
		t.Fatal(r)
	}
}
