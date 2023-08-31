package game

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/game/scheduler"
	"github.com/ethereum-optimism/optimism/op-service/clock"
	"github.com/ethereum-optimism/optimism/op-service/eth"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
)

type blockNumberFetcher func(ctx context.Context) (uint64, error)

// gameSource loads information about the games available to play
type gameSource interface {
	FetchAllGamesAtBlock(ctx context.Context, earliest uint64, blockNumber *big.Int) ([]FaultDisputeGame, error)
}

type gameScheduler interface {
	Schedule([]common.Address) error
}

type gameMonitor struct {
	logger           log.Logger
	clock            clock.Clock
	source           gameSource
	scheduler        gameScheduler
	gameWindow       time.Duration
	fetchBlockNumber blockNumberFetcher
	allowedGames     []common.Address
	l1HeadsSub       ethereum.Subscription
	l1Source         eth.NewHeadSource
}

func newGameMonitor(
	logger log.Logger,
	cl clock.Clock,
	source gameSource,
	scheduler gameScheduler,
	gameWindow time.Duration,
	fetchBlockNumber blockNumberFetcher,
	allowedGames []common.Address,
	l1Source eth.NewHeadSource,
) *gameMonitor {
	return &gameMonitor{
		logger:           logger,
		clock:            cl,
		scheduler:        scheduler,
		source:           source,
		gameWindow:       gameWindow,
		fetchBlockNumber: fetchBlockNumber,
		allowedGames:     allowedGames,
		l1Source:         l1Source,
	}
}

func (m *gameMonitor) allowedGame(game common.Address) bool {
	if len(m.allowedGames) == 0 {
		return true
	}
	for _, allowed := range m.allowedGames {
		if allowed == game {
			return true
		}
	}
	return false
}

func (m *gameMonitor) minGameTimestamp() uint64 {
	if m.gameWindow.Seconds() == 0 {
		return 0
	}
	// time: "To compute t-d for a duration d, use t.Add(-d)."
	// https://pkg.go.dev/time#Time.Sub
	if m.clock.Now().Unix() > int64(m.gameWindow.Seconds()) {
		return uint64(m.clock.Now().Add(-m.gameWindow).Unix())
	}
	return 0
}

func (m *gameMonitor) progressGames(ctx context.Context, blockNum uint64) error {
	games, err := m.source.FetchAllGamesAtBlock(ctx, m.minGameTimestamp(), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return fmt.Errorf("failed to load games: %w", err)
	}
	var gamesToPlay []common.Address
	for _, game := range games {
		if !m.allowedGame(game.Proxy) {
			m.logger.Debug("Skipping game not on allow list", "game", game.Proxy)
			continue
		}
		gamesToPlay = append(gamesToPlay, game.Proxy)
	}
	if err := m.scheduler.Schedule(gamesToPlay); errors.Is(err, scheduler.ErrBusy) {
		m.logger.Info("Scheduler still busy with previous update")
	} else if err != nil {
		return fmt.Errorf("failed to schedule games: %w", err)
	}
	return nil
}

func (m *gameMonitor) MonitorGames(ctx context.Context) {
	m.logger.Info("Monitoring fault dispute games")

	onNewL1Head := func(ctx context.Context, sig eth.L1BlockRef) {
		if err := m.progressGames(ctx, sig.Number); err != nil {
			m.logger.Error("Failed to progress games", "err", err)
		}
	}

	resubFn := func(innerCtx context.Context, err error) (event.Subscription, error) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err != nil {
			m.logger.Warn("resubscribing after failed L1 subscription", "err", err)
		}
		return eth.WatchHeadChanges(ctx, m.l1Source, onNewL1Head)
	}

	m.l1HeadsSub = event.ResubscribeErr(10*time.Second, resubFn)
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-m.l1HeadsSub.Err():
			if !ok {
				return
			}
			m.logger.Error("l1 heads subscription error", "err", err)
		}
	}
}
