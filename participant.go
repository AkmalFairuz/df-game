package game

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"sync/atomic"
)

type Participant struct {
	name   string
	xuid   string
	h      *world.EntityHandle
	state  ParticipantState
	closed atomic.Bool

	voteMapIndex *int

	impl ParticipantImpl
}

// State returns the state of the participant.
func (par *Participant) State() ParticipantState {
	return par.state
}

// Player returns the player of the participant. If the participant is not a player, the second return value is false.
func (par *Participant) Player(tx *world.Tx) (*player.Player, bool) {
	p, ok := par.h.Entity(tx)
	if !ok {
		return nil, false
	}
	return p.(*player.Player), true
}

// MustPlayer returns the player of the participant. If the participant is not a player, this function panics.
func (par *Participant) MustPlayer(tx *world.Tx) *player.Player {
	p, ok := par.Player(tx)
	if !ok {
		panic("participant is not a player")
	}
	return p
}

// Impl returns the implementation of the participant.
func (par *Participant) Impl() ParticipantImpl {
	return par.impl
}

// Closed returns whether the participant is closed.
func (par *Participant) Closed() bool {
	return par.closed.Load()
}

// close closes the participant.
func (par *Participant) close() {
	if par.closed.CompareAndSwap(false, true) {
		return
	}
	_ = par.impl.Close()
	par.impl = nil
	par.h = nil
}

// Name returns the name of the participant.
func (par *Participant) Name() string {
	return par.name
}

// XUID returns the XUID of the participant.
func (par *Participant) XUID() string {
	return par.xuid
}
