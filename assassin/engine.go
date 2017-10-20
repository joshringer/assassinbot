package assassin

import (
	"strings"
	"time"
)

// GameActionConst represent various actions that occur during a game.
type GameActionConst int

const (
	// QuitAction triggers a game to end.
	QuitAction GameActionConst = iota
	// AttackAction occurs when a player has attacked another.
	AttackAction
)

type actionQueue struct {
	q []struct {
		p, t ID
		r    bool
	}
}

func newActionQueue() *actionQueue {
	var a = new(actionQueue)
	a.q = make([]struct {
		p, t ID
		r    bool
	}, 0, 1)
	return a
}

func (a *actionQueue) each(f func(p, t ID, r bool) bool) {
	for i, v := range a.q {
		a.q[i].r = f(v.p, v.t, v.r)
	}
}

func (a *actionQueue) push(p, t ID) {
	a.q = append(a.q, struct {
		p, t ID
		r    bool
	}{p, t, false})
}

func (a *actionQueue) pop() (p, t ID, r bool) {
	var i = a.q[0]
	a.q = a.q[1:]
	return i.p, i.t, i.r
}

// GameEngine contains state information for running a game.
type GameEngine struct {
	tpl     Lang
	msg     MessageHandler
	atf     func(...int) time.Duration
	running bool
	talk    chan struct {
		ID
		string
	}
	action chan GameActionConst
}

/*
MessageHandler interface for the GameEngine to report events to.
	Announce sends a public message to all players in the game.
	Notify sends a private message to an individual player.
*/
type MessageHandler interface {
	Announce(s string)
	Notify(p Player, s string)
}

// NewGameEngine returns a new GameEngine instance.
func NewGameEngine(tpl Lang, msg MessageHandler, atf func(...int) time.Duration) *GameEngine {
	var e = new(GameEngine)
	e.tpl = tpl
	e.msg = msg
	e.atf = atf
	e.running = false
	e.talk = make(chan struct {
		ID
		string
	})
	e.action = make(chan GameActionConst)
	return e
}

func (e *GameEngine) notifyStatus(p Player) {
	var s string
	if p.Alive {
		if t, ok := p.GetTarget(); ok {
			s = e.tpl.Fmt(e.tpl.PT, t.Name, p.KillWord)
		} else {
			s = e.tpl.PA
		}
	} else {
		s = e.tpl.PD
	}
	e.msg.Notify(p, s)
}

type GameInProgressError struct {
	tpl Lang
}

func (e GameInProgressError) Error() string { return e.tpl.EIP }

func (e *GameEngine) Run(g *Game) error {
	if e.running {
		return &GameInProgressError{e.tpl}
	}
	e.running = true
	e.msg.Announce(e.tpl.GS)
	g.Start()
	g.WithPlayers(func(p Player) {
		e.notifyStatus(p)
	})
	// the main event loop
	var pc = g.Status()
	var attacks = newActionQueue()
	var elimination = func(p Player) {
		e.msg.Announce(e.tpl.Fmt(e.tpl.GD, p.Name))
		if c, ok := p.GetContract(); ok {
			e.notifyStatus(c)
		}
		pc--
	}
	for pc > 1 {
		select {
		case chat := <-e.talk:
			var p, ok = g.GetPlayer(chat.ID)
			if ok && p.Alive {
				if c, ok := p.GetContract(); ok && strings.Contains(chat.string, c.KillWord) {
					// p assassinated
					if k, ok := g.ResolvePlayerKill(p.ID); ok {
						elimination(k)
					}
				} else if strings.Contains(chat.string, p.KillWord) {
					var retaliated = false
					attacks.each(func(ap, at ID, r bool) bool {
						if at == p.ID && !r {
							// p is retaliating
							retaliated = true
							return true
						}
						return false

					})
					if !retaliated {
						// p is attacking
						if t, ok := p.GetTarget(); ok {
							attacks.push(p.ID, t.ID)
							go func() {
								time.Sleep(e.atf())
								e.action <- AttackAction
							}()
						}
					}
				}
			}
		case a := <-e.action:
			switch a {
			case QuitAction:
				pc = 0
			case AttackAction:
				var pid, tid, r = attacks.pop()
				if r {
					if k, ok := g.ResolvePlayerCounter(tid, pid); ok {
						if t, ok := g.GetPlayer(tid); ok {
							e.msg.Notify(t, e.tpl.PCS)
						}
						elimination(k)
					}
				} else {
					if k, ok := g.ResolvePlayerAttack(pid, tid); ok {
						if p, ok := g.GetPlayer(pid); ok {
							e.msg.Notify(p, e.tpl.PAS)
						}
						elimination(k)
					}
				}
			}
		}
	}
	e.msg.Announce(e.tpl.GE)
	var w = make([]string, 0, 1)
	g.WithPlayers(func(p Player) {
		if p.Alive {
			w = append(w, p.Name)
		}
	})
	if len(w) == 1 {
		e.msg.Announce(e.tpl.Fmt(e.tpl.GW, w[0]))
	} else {
		e.msg.Announce(e.tpl.Fmt(e.tpl.GWM, w))
	}
	e.running = false
	return nil
}

func (e *GameEngine) IncomingTalk(from ID, text string) {
	e.talk <- struct {
		ID
		string
	}{from, text}
}
