package game

import (
	"fmt"
	form "github.com/akmalfairuz/ez-form"
	"github.com/df-mc/dragonfly/server/player"
	"strings"
)

// Teleporter is a game feature that allows players to teleport to other players.
func sendTeleporterForm(g *Game, p *player.Player) {
	m := form.NewMenu("Teleporter")
	m.WithContent("Select a player to teleport to:")
	xuids := make([]string, 0)
	for participant := range g.Participants() {
		if !participant.state.Playing() {
			continue
		}
		p2 := participant.MustPlayer(p.Tx())
		m.WithButton(p2.Name(), fmt.Sprintf("https://player.venitymc.com/%s/avatar.png", strings.ToLower(p2.Name())))
		xuids = append(xuids, p2.XUID())
	}

	m.WithCallback(func(p *player.Player, result int) {
		if g.closed.Load() {
			return
		}
		if !g.InGame(p) {
			return
		}
		targetPar, ok := g.ParticipantByXUID(xuids[result])
		if !ok {
			return
		}

		p.Teleport(targetPar.MustPlayer(p.Tx()).Position())
	})

	p.SendForm(m)
}
