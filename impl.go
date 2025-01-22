package game

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

type Impl interface {
	// MaxPlayers returns the maximum amount of players that can be in the game at once.
	MaxPlayers() int
	// MinPlayers returns the minimum amount of players that are required to start the game.
	MinPlayers() int
	// WaitingDuration returns the duration that the game should wait before starting after the minimum amount of players
	// has joined.
	WaitingDuration() time.Duration
	// HandleStart is called when the game should start.
	HandleStart(tx *world.Tx)
	// HandleQuit is called when a player quits the game.
	HandleQuit(tx *world.Tx, p *Participant)
	// HandleJoin is called when a player joins the game.
	HandleJoin(tx *world.Tx, p *Participant)
	// HandleClose is called when the game is closed.
	HandleClose(tx *world.Tx)
	// HandlePlayingTick is called every tick while the game is playing.
	HandlePlayingTick(tx *world.Tx, currentTick uint64)
	// RenderWaitingScoreboard renders the scoreboard for the waiting State.
	RenderWaitingScoreboard(p *player.Player, startingIn int, participantLen int)
	// RenderFinishedScoreboard renders the scoreboard for the finished State.
	RenderFinishedScoreboard(p *player.Player, closingIn int)
	// HandleMapReady is called when the map is ready to be played.
	HandleMapReady(tx *world.Tx, m *Map)
	// Load is called when the game is loaded.
	Load()
	// HandleParticipantCreate is called when a participant is created. It should return the implementation of the participant.
	HandleParticipantCreate(par *Participant) ParticipantImpl
}

type Allower interface {
	Allow(p *player.Player) (string, bool)
}

type SetterGame interface {
	// SetGame sets the game that the player is in.
	SetGame(g *Game)
}

type ParticipantImpl interface {
	Close() error
}
