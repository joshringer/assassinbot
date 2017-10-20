package assassin

import (
	"math/rand"
	"time"
)

// ID == identifier, used to uniquely identify Players/Games.
type ID int

// Player contains player state information.
type Player struct {
	ID
	Name     string
	Alive    bool
	kwg      WordGenerator
	KillWord string
	target   *Player
	contract *Player
}

// NewPlayer creates a new Player instance.
func NewPlayer(id ID, name string, kwg WordGenerator) *Player {
	return &Player{
		ID:    id,
		Name:  name,
		Alive: true,
		kwg:   kwg,
	}
}

// GetTarget retrieves player target info.
func (p Player) GetTarget() (t Player, ok bool) {
	ok = p.target != nil
	if ok {
		t = *(p.target)
	}
	return
}

// GetContract retrieves player contract info.
func (p Player) GetContract() (c Player, ok bool) {
	ok = p.contract != nil
	if ok {
		c = *(p.contract)
	}
	return
}

type PlayerDeadError struct {
	string
}

func (e PlayerDeadError) Error() string {
	return e.string
}

// SetTarget sets player target as provided.
func (p *Player) SetTarget(t *Player) error {
	if !p.Alive || t != nil && !t.Alive {
		return &PlayerDeadError{"Both player and target must be alive"}
	}
	if p.target != nil && p.target.Alive {
		p.target.contract = nil
	}
	p.target = t
	if t == nil {
		p.KillWord = ""
	} else {
		p.KillWord = p.kwg.Gen()
		t.contract = p
	}
	return nil
}

// SetEliminated sets player status to dead and reassigns the contract for them to their target.
func (p *Player) SetEliminated() error {
	if !p.Alive {
		return &PlayerDeadError{"Player is already dead"}
	}
	p.Alive = false
	p.contract.SetTarget(p.target)
	return nil
}

// Game contains game state information.
type Game struct {
	ID
	players map[ID]*Player
}

// NewGame creates a new Game instance.
func NewGame(id ID, playerList map[ID]string, kwg WordGenerator) *Game {
	var g = &Game{
		ID:      id,
		players: make(map[ID]*Player, len(playerList))}
	for id, name := range playerList {
		g.players[id] = NewPlayer(id, name, kwg)
	}
	return g
}

// GetPlayer retreives player info for given id.
func (g Game) GetPlayer(id ID) (p Player, ok bool) {
	pp, ok := g.players[id]
	if ok {
		p = *pp
	}
	return
}

// WithPlayers loops over all players in game (alive and dead).
func (g Game) WithPlayers(f func(p Player)) {
	for _, p := range g.players {
		f(*p)
	}
}

/*
ResolvePlayerKill action in game: player with id killed.
If player was alive, return killed player detail and ok if action successful.
*/
func (g *Game) ResolvePlayerKill(id ID) (Player, bool) {
	if p, ok := g.players[id]; ok && p.Alive {
		p.SetEliminated()
		return *p, true
	}
	return Player{}, false
}

/*
ResolvePlayerAttack action in game: pid attacked tid.
Return killed player info, ok if attack successful.
*/
func (g *Game) ResolvePlayerAttack(pid, tid ID) (Player, bool) {
	var p, pok = g.players[pid]
	var t, tok = g.players[tid]
	if pok && tok && p.Alive && t.Alive && p.target == t {
		t.SetEliminated()
		return *t, true
	}
	return Player{}, false
}

/*
ResolvePlayerCounter action in game: pid countered cid.
Return killed player info, ok if counter successful.
*/
func (g *Game) ResolvePlayerCounter(pid, cid ID) (Player, bool) {
	var p, pok = g.players[pid]
	var c, cok = g.players[cid]
	if pok && cok && p.Alive && c.Alive && p.contract == c {
		c.SetEliminated()
		return *c, true
	}
	return Player{}, false
}

// Start starts the game, assigning player targets.
func (g *Game) Start() {
	var (
		l  = len(g.players)
		il = rand.Perm(l)
		pl = make([]*Player, 0, l)
	)
	for _, p := range g.players {
		pl = append(pl, p)
	}
	for i := 0; i < len(g.players); i++ {
		pl[il[i]].SetTarget(pl[il[(i+1)%l]])
	}
}

// Status returns count of players still alive in the game.
func (g Game) Status() int {
	var c = 0
	for _, p := range g.players {
		if p.Alive {
			c++
		}
	}
	return c
}

func init() {
	rand.Seed(time.Now().Unix())
}
