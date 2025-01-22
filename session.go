package game

import (
	"github.com/akmalfairuz/df-game/internal"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"sync"
)

var globalSessionManager = internal.NewMap[string, *Session]()

func GlobalSessionManager() *internal.Map[string, *Session] {
	return globalSessionManager
}

type Session struct {
	mu sync.Mutex

	xuid string
	name string
	g    *Game
	h    *world.EntityHandle
}

// NewSession creates a new session for the player p.
func NewSession(p *player.Player) *Session {
	return &Session{
		xuid: p.XUID(),
		name: p.Name(),
		h:    p.H(),
	}
}

// XUID returns the XUID of the player that the session is for.
func (s *Session) XUID() string {
	return s.xuid
}

// Name returns the name of the player that the session is for.
func (s *Session) Name() string {
	return s.name
}

// Game returns the game that the session is in, if any.
func (s *Session) Game() (*Game, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g := s.g
	if g != nil && g.closed.Load() {
		return nil, false
	}
	return g, g != nil
}

// SetGame sets the game that the session is in.
func (s *Session) SetGame(g *Game) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.g != nil && g != nil {
		panic("set non nil game is not allowed if the session already has a game")
	}
	s.g = g
}

// EntityHandle returns the entity handle of the player that the session is for.
func (s *Session) EntityHandle() *world.EntityHandle {
	return s.h
}

// Player returns the player that the session is for.
func (s *Session) Player(tx *world.Tx) (*player.Player, bool) {
	p, ok := s.h.Entity(tx)
	if !ok {
		return nil, false
	}
	return p.(*player.Player), true
}

// Close closes the session.
func (s *Session) Close() {
	s.g = nil
	s.h = nil
}
