package game

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"log/slog"
)

type Config struct {
	ID            uuid.UUID
	Log           *slog.Logger
	Impl          Impl
	MapsDir       string
	PlayerHandler player.Handler
	WorldHandler  world.Handler
	PlayAgainHook func(p *player.Player)
}

var DefaultWaitingWorld *world.World

func (c *Config) New() (*Game, error) {
	g := &Game{
		log:           c.Log,
		id:            c.ID,
		mDir:          c.MapsDir,
		impl:          c.Impl,
		ph:            c.PlayerHandler,
		wh:            c.WorldHandler,
		playAgainHook: c.PlayAgainHook,
	}
	if err := g.Load(); err != nil {
		return nil, err
	}
	return g, nil
}
