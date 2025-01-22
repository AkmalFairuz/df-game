package game

import (
	"fmt"
	"github.com/akmalfairuz/df-game/internal"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/google/uuid"
	"iter"
)

// Factory is a factory for games.
type Factory struct {
	games      *internal.Map[uuid.UUID, *Game]
	createFunc func() *Game
}

// FactoryConfig is a configuration for a game factory.
type FactoryConfig struct {
	CreateFunc func() *Game
}

// New creates a new game factory.
func (c FactoryConfig) New() *Factory {
	return &Factory{
		games:      internal.NewMap[uuid.UUID, *Game](),
		createFunc: c.CreateFunc,
	}
}

// Games returns available games.
func (f *Factory) Games() iter.Seq[*Game] {
	return func(yield func(*Game) bool) {
		for _, g := range f.games.Map() {
			if !yield(g) {
				break
			}
		}
	}
}

// Join joins a player to a game.
func (f *Factory) Join(p *player.Player) (*Game, bool) {
	for g := range f.Games() {
		if err := g.Join(p); err == nil {
			return g, true
		}
	}

	newG := f.NewGame()
	if err := newG.Join(p); err != nil {
		fmt.Println(err)
		return nil, false
	}
	return newG, true
}

// NewGame creates a new game.
func (f *Factory) NewGame() *Game {
	g := (f.createFunc)()
	g.closeHook = func() {
		f.games.Delete(g.ID())
	}
	f.games.Store(g.ID(), g)
	return g
}
