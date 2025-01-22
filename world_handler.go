package game

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type worldHandler struct {
	g *Game
}

// worldHandler ...
var _ world.Handler = &worldHandler{}

func (wh *worldHandler) HandleLiquidFlow(ctx *world.Context, from, into cube.Pos, liquid world.Liquid, replaced world.Block) {
	if !wh.g.State().Playing() {
		ctx.Cancel()
		return
	}
}

func (wh *worldHandler) HandleLiquidDecay(ctx *world.Context, pos cube.Pos, before, after world.Liquid) {
	if !wh.g.State().Playing() {
		ctx.Cancel()
		return
	}
}

func (wh *worldHandler) HandleLiquidHarden(ctx *world.Context, hardenedPos cube.Pos, liquidHardened, otherLiquid, newBlock world.Block) {
	if !wh.g.State().Playing() {
		ctx.Cancel()
		return
	}

	wh.g.wh.HandleLiquidHarden(ctx, hardenedPos, liquidHardened, otherLiquid, newBlock)
}

func (wh *worldHandler) HandleSound(ctx *world.Context, s world.Sound, pos mgl64.Vec3) {
	wh.g.wh.HandleSound(ctx, s, pos)
}

func (wh *worldHandler) HandleFireSpread(ctx *world.Context, from, to cube.Pos) {
	if !wh.g.State().Playing() {
		ctx.Cancel()
		return
	}

	wh.g.wh.HandleFireSpread(ctx, from, to)
}

func (wh *worldHandler) HandleBlockBurn(ctx *world.Context, pos cube.Pos) {
	if !wh.g.State().Playing() {
		ctx.Cancel()
		return
	}

	wh.g.wh.HandleBlockBurn(ctx, pos)
}

func (wh *worldHandler) HandleCropTrample(ctx *world.Context, pos cube.Pos) {
	if !wh.g.State().Playing() {
		ctx.Cancel()
		return
	}

	wh.g.wh.HandleCropTrample(ctx, pos)
}

func (wh *worldHandler) HandleLeavesDecay(ctx *world.Context, pos cube.Pos) {
	ctx.Cancel()
	wh.g.wh.HandleLeavesDecay(ctx, pos)
}

func (wh *worldHandler) HandleEntitySpawn(tx *world.Tx, e world.Entity) {
	wh.g.wh.HandleEntitySpawn(tx, e)
}

func (wh *worldHandler) HandleEntityDespawn(tx *world.Tx, e world.Entity) {
	wh.g.wh.HandleEntityDespawn(tx, e)
}

func (wh *worldHandler) HandleClose(tx *world.Tx) {
	wh.g.wh.HandleClose(tx)
}
