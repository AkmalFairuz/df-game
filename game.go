package game

import (
	"errors"
	"fmt"
	"github.com/akmalfairuz/df-game/internal"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"iter"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Game ...
type Game struct {
	log *slog.Logger

	id           uuid.UUID
	participants *internal.Map[string, *Participant]

	s   State
	sMu sync.Mutex

	w *world.World

	m             *Map
	mDir          string
	availableMaps []*Map

	tickQueue chan struct{}

	impl Impl

	closed atomic.Bool

	startingIn  int
	closingIn   int
	currentTick atomic.Uint64

	mapLoaded bool
	wPath     string

	playAgainHook func(p *player.Player)

	closeHook func()

	ph player.Handler
	wh world.Handler
}

var (
	gameItemKey         = "gameItem"
	quitItemValue       = "quit"
	playAgainItemValue  = "playAgain"
	teleporterItemValue = "teleporter"
	voteMapItemValue    = "voteMap"

	quitItem       = item.NewStack(item.DragonBreath{}, 1).WithCustomName(text.Colourf("<red>Quit</red>")).WithValue(gameItemKey, quitItemValue)
	playAgainItem  = item.NewStack(item.Paper{}, 1).WithCustomName(text.Colourf("<green>Play Again</green>")).WithValue(gameItemKey, playAgainItemValue)
	teleporterItem = item.NewStack(item.Compass{}, 1).WithCustomName(text.Colourf("<yellow>Teleporter</yellow>")).WithValue(gameItemKey, teleporterItemValue)
	voteMapItem    = item.NewStack(item.Paper{}, 1).WithCustomName(text.Colourf("<yellow>Vote Map</yellow>")).WithValue(gameItemKey, voteMapItemValue)
)

// ID returns the ID of the game.
func (g *Game) ID() uuid.UUID {
	return g.id
}

// Load ...
func (g *Game) Load() error {
	maps, err := loadMaps(g.mDir)
	if err != nil {
		return fmt.Errorf("failed to load maps: %w", err)
	}

	if g.id == uuid.Nil {
		g.id = uuid.New()
	}
	g.closed.Store(false)
	g.availableMaps = maps
	g.participants = internal.NewMap[string, *Participant]()
	g.tickQueue = make(chan struct{}, 32)
	g.setState(StateWaiting)
	g.startingIn = int(g.impl.WaitingDuration().Seconds())
	g.wPath = filepath.Join("game_worlds", g.id.String())
	g.impl.Load()

	if g.wh == nil {
		g.wh = world.NopHandler{}
	}
	if g.ph == nil {
		g.ph = player.NopHandler{}
	}

	go g.startTicking()
	return nil
}

// MapLoaded returns whether the map is loaded.
func (g *Game) MapLoaded() bool {
	return g.mapLoaded
}

// Map returns the map that the game is currently using.
func (g *Game) Map() *Map {
	if g.m == nil {
		panic("map not loaded")
	}
	return g.m
}

// State returns the current State of the game.
func (g *Game) State() State {
	g.sMu.Lock()
	defer g.sMu.Unlock()
	return g.s
}

// setState sets the State of the game to the State passed.
func (g *Game) setState(s State) {
	g.sMu.Lock()
	g.s = s
	g.sMu.Unlock()
}

// startTicking is used to start the game ticking. This function should be called in a goroutine.
func (g *Game) startTicking() {
	t := time.NewTicker(time.Second / 20)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if g.closed.Load() {
				return
			}
			g.world().Exec(g.onTick)
		}
	}
}

// Impl returns the implementation of the game.
func (g *Game) Impl() Impl {
	return g.impl
}

// onTick is called every tick of the game.
func (g *Game) onTick(tx *world.Tx) {
	if g.closed.Load() || !g.ValidTx(tx) {
		return
	}
	g.currentTick.Add(1)
	currentTick := g.currentTick.Load()

	switch g.State() {
	case StateWaiting:
		if currentTick%20 != 0 {
			break
		}
		enoughPlayers := g.participants.Len() >= g.impl.MinPlayers()
		if enoughPlayers {
			g.startingIn--
			if g.startingIn <= 4 && !g.mapLoaded {
				if err := g.loadMap(tx); err != nil {
					panic(err)
				}
			}
			if g.startingIn <= 0 {
				g.Start(tx)
			}
		} else {
			g.startingIn = int(g.impl.WaitingDuration().Seconds())
		}

		participantLen := g.participants.Len()
		s := -1
		if enoughPlayers {
			s = g.startingIn
		}
		g.Players(tx, func(p *player.Player, par *Participant) {
			g.impl.RenderWaitingScoreboard(p, s, participantLen)
		})
	case StatePlaying:
		g.impl.HandlePlayingTick(tx, currentTick)
	case StateFinished:
		if currentTick%20 != 0 {
			break
		}
		g.closingIn--
		if g.closingIn <= 0 {
			g.close(tx)
			break
		}

		g.Players(tx, func(p *player.Player, par *Participant) {
			g.impl.RenderFinishedScoreboard(p, g.closingIn)
		})
	default: // unknown State
	}
}

// ValidTx returns whether the transaction passed is valid for the game.
func (g *Game) ValidTx(tx *world.Tx) (valid bool) {
	defer func() {
		if r := recover(); r != nil {
			valid = false
			g.log.Error("recovered from panic", "panic", r)
		}
	}()

	return tx.World() == g.world()
}

// world returns the world that the game is currently in.
func (g *Game) world() *world.World {
	if g.State().Waiting() {
		return DefaultWaitingWorld
	}
	if g.w != nil {
		return g.w
	}
	panic("world not loaded")
}

// Join is used to join a player to the game. If the game is not in the waiting State, an error is returned.
func (g *Game) Join(p *player.Player) error {
	if g.closed.Load() {
		return errors.New("game is closed")
	}

	if !g.State().Waiting() {
		return errors.New("game is not in waiting State")
	}

	if !g.ValidTx(p.Tx()) {
		return errors.New("invalid tx: expected player to be in the waiting world")
	}

	if g.participants.Len() >= g.impl.MaxPlayers() {
		return errors.New("game is full")
	}

	sess, ok := globalSessionManager.Load(p.XUID())
	if !ok {
		return errors.New("player session not found")
	}

	if _, ok := sess.Game(); ok {
		return errors.New("player is already in a game")
	}

	sess.SetGame(g)
	spawnPos := DefaultWaitingWorld.Spawn().Vec3Middle()
	p.Teleport(spawnPos)
	p.Messagef("Teleported to %.1f, %.1f, %.1f", spawnPos.X(), spawnPos.Y(), spawnPos.Z())

	if allower, ok := g.impl.(Allower); ok {
		reason, allowed := allower.Allow(p)
		if !allowed {
			return fmt.Errorf("player not allowed to join: %s", reason)
		}
	}

	par := &Participant{
		name:  p.Name(),
		xuid:  p.XUID(),
		h:     p.H(),
		state: ParticipantStatePlaying,
	}

	par.impl = g.impl.HandleParticipantCreate(par)
	g.participants.Store(p.XUID(), par)

	resetPlayer(p)
	p.SetGameMode(world.GameModeAdventure)
	_ = p.Inventory().SetItem(0, voteMapItem)
	_ = p.Inventory().SetItem(8, quitItem)

	for e := range p.Tx().Players() {
		if e.H() == p.H() {
			continue
		}

		eP := e.(*player.Player)
		if !g.InGame(eP) {
			p.HideEntity(e)
			eP.HideEntity(p)
		} else {
			p.ShowEntity(e)
			eP.ShowEntity(p)
		}
	}

	g.impl.HandleJoin(p.Tx(), par)

	return nil
}

// Participants returns the participants in the game.
func (g *Game) Participants() iter.Seq[*Participant] {
	return func(yield func(*Participant) bool) {
		for _, par := range g.participants.Map() {
			if !yield(par) {
				return
			}
		}
	}
}

// ParticipantLen returns the amount of participants in the game.
func (g *Game) ParticipantLen() int {
	return g.participants.Len()
}

// PlayingParticipantLen returns the amount of participants in the game that are playing.
func (g *Game) PlayingParticipantLen() int {
	var n int
	for _, par := range g.participants.Map() {
		if par.state.Playing() {
			n++
		}
	}
	return n
}

// PlayingParticipants returns the participants in the game that are playing.
func (g *Game) PlayingParticipants() iter.Seq[*Participant] {
	return func(yield func(*Participant) bool) {
		for _, par := range g.participants.Map() {
			if par.state.Playing() {
				if !yield(par) {
					return
				}
			}
		}
	}
}

// ParticipantByXUID returns the participant with the XUID passed.
func (g *Game) ParticipantByXUID(xuid string) (*Participant, bool) {
	return g.participants.Load(xuid)
}

// InGame returns whether the player passed is in the game.
func (g *Game) InGame(p *player.Player) bool {
	_, ok := g.participants.Load(p.XUID())
	return ok
}

// Leave is used to remove a player from the game.
func (g *Game) Leave(p *player.Player) (bool, error) {
	if !g.ValidTx(p.Tx()) {
		return false, errors.New("expected player to be in the world of the game")
	}

	sess, ok := globalSessionManager.Load(p.XUID())
	if !ok {
		return false, errors.New("player session not found")
	}

	currentG, ok := sess.Game()
	if !ok {
		return false, errors.New("player is not in a game")
	}

	if currentG != g {
		return false, errors.New("player is not in this game")
	}

	sess.SetGame(nil)

	par, ok := g.participants.Load(p.XUID())
	if !ok {
		return false, errors.New("player is not in the game")
	}

	g.impl.HandleQuit(p.Tx(), par)
	resetPlayer(p)

	g.participants.Delete(p.XUID())

	worldChanged := false
	if !g.State().Waiting() {
		h := p.Tx().RemoveEntity(p)
		<-DefaultWaitingWorld.Exec(func(newTx *world.Tx) {
			newP := newTx.AddEntity(h).(*player.Player)
			newP.Teleport(DefaultWaitingWorld.Spawn().Vec3Middle())
			for e := range newTx.Players() {
				if e.H() == newP.H() {
					continue
				}
				e.(*player.Player).HideEntity(newP)
				newP.HideEntity(e)
			}
		})
		worldChanged = true
	}
	par.close()

	return worldChanged, nil
}

// Players are used to iterate over all players in the game, calling the function passed for each player.
func (g *Game) Players(tx *world.Tx, fn func(p *player.Player, par *Participant)) {
	if !g.ValidTx(tx) {
		return
	}

	for _, par := range g.participants.Map() {
		p, ok := par.Player(tx)
		if !ok {
			continue
		}

		fn(p, par)
	}
}

// PlayingPlayers are used to iterate over all players in the game that are playing, calling the function passed
// for each player.
func (g *Game) PlayingPlayers(tx *world.Tx, fn func(p *player.Player, par *Participant)) {
	if !g.ValidTx(tx) {
		return
	}

	for _, par := range g.participants.Map() {
		if !par.state.Playing() {
			continue
		}

		p, ok := par.Player(tx)
		if !ok {
			continue
		}

		fn(p, par)
	}
}

// Messagef is used to send a message to all players in the game.
func (g *Game) Messagef(tx *world.Tx, format string, args ...any) {
	g.Players(tx, func(p *player.Player, _ *Participant) {
		p.Messagef(text.Colourf(format, args...))
	})
}

// loadMap ...
func (g *Game) loadMap(tx *world.Tx) error {
	if !g.ValidTx(tx) {
		return errors.New("expected transaction to be valid")
	}

	if g.mapLoaded {
		return errors.New("map already loaded")
	}

	votes := make([]int, len(g.availableMaps))

	g.Players(tx, func(p *player.Player, par *Participant) {
		if par.voteMapIndex == nil {
			return
		}

		votes[*par.voteMapIndex]++
	})

	maxVotes := 0
	maxVotesIndex := -1
	for i, v := range votes {
		if v > maxVotes {
			maxVotes = v
			maxVotesIndex = i
		}
	}

	if maxVotes == 0 {
		maxVotesIndex = rand.Intn(len(g.availableMaps))
	}

	selectedMap := g.availableMaps[maxVotesIndex]
	g.log.Info("selected map", "map", selectedMap.Name)

	if err := selectedMap.CopyWorldTo(g.wPath); err != nil {
		return fmt.Errorf("failed to copy map world: %w", err)
	}

	prov, err := mcdb.Open(g.wPath)
	if err != nil {
		return fmt.Errorf("failed to open world: %w", err)
	}

	wConf := world.Config{
		Dim:          world.Overworld,
		Provider:     prov,
		Generator:    world.NopGenerator{},
		ReadOnly:     false,
		SaveInterval: time.Hour,
		Entities:     entity.DefaultRegistry,
	}

	g.w = wConf.New()
	g.w.StopTime()
	g.w.SetTime(3000)
	g.w.StopThundering()
	g.w.StopRaining()
	g.w.StopWeatherCycle()
	g.w.SetDifficulty(world.DifficultyEasy)

	g.w.Handle(&worldHandler{g: g})

	g.mapLoaded = true
	g.m = selectedMap

	if v, ok := g.wh.(SetterGame); ok {
		v.SetGame(g)
	}

	if v, ok := g.ph.(SetterGame); ok {
		v.SetGame(g)
	}

	if v, ok := g.impl.(SetterGame); ok {
		v.SetGame(g)
	}

	g.impl.HandleMapReady(tx, g.m)

	return nil
}

// Start is used to start the game. If the game is not in the waiting State, this function does nothing.
func (g *Game) Start(tx *world.Tx) {
	if !g.ValidTx(tx) || !g.State().Waiting() {
		return
	}
	if !g.mapLoaded {
		if err := g.loadMap(tx); err != nil {
			panic(err)
		}
	}

	h := make([]*world.EntityHandle, 0, g.participants.Len())

	g.Players(tx, func(p *player.Player, _ *Participant) {
		resetPlayer(p)

		pH := tx.RemoveEntity(p)
		h = append(h, pH)
	})

	g.setState(StatePlaying)

	g.w.Exec(func(newTx *world.Tx) {
		for _, pH := range h {
			newTx.AddEntity(pH)
		}

		g.impl.HandleStart(newTx)
	})
}

// End is used to end the game.
func (g *Game) End(tx *world.Tx) {
	if !g.ValidTx(tx) || !g.State().Playing() {
		return
	}

	g.setState(StateFinished)
	g.closingIn = 3

	g.Players(tx, func(p *player.Player, par *Participant) {
		resetPlayer(p)

		if par.state.Playing() {
			p.SetGameMode(world.GameModeAdventure)
		}

		_ = p.SetHeldSlot(1)
		_ = p.Inventory().SetItem(0, playAgainItem)
		_ = p.Inventory().SetItem(8, quitItem)
	})
}

// close is used to close the game.
func (g *Game) close(tx *world.Tx) {
	if g.closed.Load() {
		return
	}
	if !g.ValidTx(tx) {
		g.log.Warn("expected transaction to be valid when closing game")
		return
	}

	g.impl.HandleClose(tx)

	g.Players(tx, func(p *player.Player, par *Participant) {
		g.playAgain(p)
	})

	g.setState(StateUnknown)
	g.closed.Store(true)

	if g.closeHook != nil {
		(g.closeHook)()
	}

	if g.w != nil {
		DefaultWaitingWorld.Exec(func(tx *world.Tx) {
			_ = g.w.Close()
			if err := os.RemoveAll(g.wPath); err != nil {
				g.log.Error("failed to remove world directory", "path", g.wPath, "error", err)
			}
		})
	}

	g.log.Info("game closed")
}

// SetSpectator is used to set a player to spectator mode.
func (g *Game) SetSpectator(p *player.Player) {
	if !g.ValidTx(p.Tx()) {
		return
	}

	par, ok := g.participants.Load(p.XUID())
	if !ok {
		return
	}

	par.state = ParticipantStateSpectating

	resetPlayer(p)
	p.SetGameMode(world.GameModeSpectator)
	_ = p.SetHeldSlot(1)
	_ = p.Inventory().SetItem(0, playAgainItem)
	_ = p.Inventory().SetItem(4, teleporterItem)
	_ = p.Inventory().SetItem(8, quitItem)
}

func (g *Game) playAgain(p *player.Player) {
	if !g.ValidTx(p.Tx()) || !g.InGame(p) {
		return
	}

	worldChanged, err := g.Leave(p)
	if err != nil {
		panic(err)
	}

	if worldChanged {
		<-DefaultWaitingWorld.Exec(func(tx *world.Tx) {
			newP, ok := p.H().Entity(tx)
			if !ok {
				return
			}

			(g.playAgainHook)(newP.(*player.Player))
		})
		return
	}

	(g.playAgainHook)(p)
}
