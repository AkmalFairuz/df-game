package game

import (
	form "github.com/akmalfairuz/ez-form"
	"github.com/df-mc/dragonfly/server/player"
)

func sendVoteMapForm(g *Game, p *player.Player) {
	f := form.NewMenu("Vote Map")
	f.WithContent("Select a map to vote:")
	for _, m := range g.availableMaps {
		f.WithButton(m.Name)
	}
	f.WithCallback(func(p *player.Player, result int) {
		if g.closed.Load() || !g.InGame(p) {
			return
		}
		par, ok := g.ParticipantByXUID(p.XUID())
		if !ok {
			return
		}

		par.voteMapIndex = &result
	})
	p.SendForm(f)
}
