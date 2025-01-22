package game

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"net"
	"time"
)

type PlayerHandler struct{}

// PlayerHandler ...
var _ player.Handler = &PlayerHandler{}

func phExec(p *player.Player, f func(g *Game)) {
	sess, ok := globalSessionManager.Load(p.XUID())
	if !ok {
		return
	}
	g, ok := sess.Game()
	if !ok {
		return
	}
	if !g.ValidTx(p.Tx()) {
		return
	}
	f(g)
}

func (ph *PlayerHandler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleMove(ctx, newPos, newRot)
	})
}

func (ph *PlayerHandler) HandleJump(p *player.Player) {
	phExec(p, func(g *Game) {
		g.ph.HandleJump(p)
	})
}

func (ph *PlayerHandler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleTeleport(ctx, pos)
	})
}

func (ph *PlayerHandler) HandleChangeWorld(p *player.Player, before, after *world.World) {
	phExec(p, func(g *Game) {
		g.ph.HandleChangeWorld(p, before, after)
	})
}

func (ph *PlayerHandler) HandleToggleSprint(ctx *player.Context, after bool) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleToggleSprint(ctx, after)
	})
}

func (ph *PlayerHandler) HandleToggleSneak(ctx *player.Context, after bool) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleToggleSneak(ctx, after)
	})
}

func (ph *PlayerHandler) HandleChat(ctx *player.Context, message *string) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleChat(ctx, message)
	})
}

func (ph *PlayerHandler) HandleFoodLoss(ctx *player.Context, from int, to *int) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleFoodLoss(ctx, from, to)
	})
}

func (ph *PlayerHandler) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleHeal(ctx, health, src)
	})
}

func (ph *PlayerHandler) HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleHurt(ctx, damage, immune, attackImmunity, src)
	})
}

func (ph *PlayerHandler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	phExec(p, func(g *Game) {
		g.ph.HandleDeath(p, src, keepInv)
	})
}

func (ph *PlayerHandler) HandleRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {
	phExec(p, func(g *Game) {
		g.ph.HandleRespawn(p, pos, w)
	})
}

func (ph *PlayerHandler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {
	ctx.Cancel()
}

func (ph *PlayerHandler) HandleFireExtinguish(ctx *player.Context, pos cube.Pos) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleFireExtinguish(ctx, pos)
	})
}

func (ph *PlayerHandler) HandleStartBreak(ctx *player.Context, pos cube.Pos) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleStartBreak(ctx, pos)
	})
}

func (ph *PlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleBlockBreak(ctx, pos, drops, xp)
	})
}

func (ph *PlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleBlockPlace(ctx, pos, b)
	})
}

func (ph *PlayerHandler) HandleBlockPick(ctx *player.Context, pos cube.Pos, b world.Block) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleBlockPick(ctx, pos, b)
	})
}

func (ph *PlayerHandler) HandleItemUse(ctx *player.Context) {
	phExec(ctx.Val(), func(g *Game) {
		heldItem, _ := ctx.Val().HeldItems()
		val, ok := heldItem.Value("gameItem")
		if ok {
			ctx.Cancel()
			switch val.(string) {
			case voteMapItemValue:
				sendVoteMapForm(g, ctx.Val())
				return
			case quitItemValue:
				_, _ = g.Leave(ctx.Val())
				return
			case playAgainItemValue:
				g.playAgain(ctx.Val())
				return
			case teleporterItemValue:
				sendTeleporterForm(g, ctx.Val())
				return
			}
		}

		g.ph.HandleItemUse(ctx)
	})
}

func (ph *PlayerHandler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleItemUseOnBlock(ctx, pos, face, clickPos)
	})
}

func (ph *PlayerHandler) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleItemUseOnEntity(ctx, e)
	})
}

func (ph *PlayerHandler) HandleItemRelease(ctx *player.Context, item item.Stack, dur time.Duration) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleItemRelease(ctx, item, dur)
	})
}

func (ph *PlayerHandler) HandleItemConsume(ctx *player.Context, item item.Stack) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleItemConsume(ctx, item)
	})
}

func (ph *PlayerHandler) HandleAttackEntity(ctx *player.Context, e world.Entity, force, height *float64, critical *bool) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleAttackEntity(ctx, e, force, height, critical)
	})
}

func (ph *PlayerHandler) HandleExperienceGain(ctx *player.Context, amount *int) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleExperienceGain(ctx, amount)
	})
}

func (ph *PlayerHandler) HandlePunchAir(ctx *player.Context) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandlePunchAir(ctx)
	})
}

func (ph *PlayerHandler) HandleSignEdit(ctx *player.Context, pos cube.Pos, frontSide bool, oldText, newText string) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleSignEdit(ctx, pos, frontSide, oldText, newText)
	})
}

func (ph *PlayerHandler) HandleLecternPageTurn(ctx *player.Context, pos cube.Pos, oldPage int, newPage *int) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleLecternPageTurn(ctx, pos, oldPage, newPage)
	})
}

func (ph *PlayerHandler) HandleItemDamage(ctx *player.Context, i item.Stack, damage int) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleItemDamage(ctx, i, damage)
	})
}

func (ph *PlayerHandler) HandleItemPickup(ctx *player.Context, i *item.Stack) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleItemPickup(ctx, i)
	})
}

func (ph *PlayerHandler) HandleHeldSlotChange(ctx *player.Context, from, to int) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleHeldSlotChange(ctx, from, to)
	})
}

func (ph *PlayerHandler) HandleItemDrop(ctx *player.Context, s item.Stack) {
	phExec(ctx.Val(), func(g *Game) {
		if !g.State().Playing() {
			ctx.Cancel()
			return
		}
		g.ph.HandleItemDrop(ctx, s)
	})
}

func (ph *PlayerHandler) HandleTransfer(ctx *player.Context, addr *net.UDPAddr) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleTransfer(ctx, addr)
	})
}

func (ph *PlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	phExec(ctx.Val(), func(g *Game) {
		g.ph.HandleCommandExecution(ctx, command, args)
	})
}

func (ph *PlayerHandler) HandleQuit(p *player.Player) {
	phExec(p, func(g *Game) {
		_, _ = g.Leave(p)
	})

	sess, ok := globalSessionManager.Load(p.XUID())
	if ok {
		sess.Close()
		globalSessionManager.Delete(p.XUID())
	}
}

func (ph *PlayerHandler) HandleDiagnostics(p *player.Player, d session.Diagnostics) {
	phExec(p, func(g *Game) {
		g.ph.HandleDiagnostics(p, d)
	})
}
