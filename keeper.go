// Package keeper is the Nectar keeper SDK: a small framework for building
// Soroban liquidation/automation keepers. Register one or more ProtocolAdapters
// and call Run; the keeper polls each adapter for tasks each cycle and executes
// them using shared vault capital.
//
// See github.com/Nectar-Network/keeper-sdk/adapters/blend for a reference
// adapter and examples/ for runnable keepers.
package keeper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/stellar/go/keypair"

	"github.com/Nectar-Network/keeper-sdk/adapters"
	"github.com/Nectar-Network/keeper-sdk/dex"
	"github.com/Nectar-Network/keeper-sdk/soroban"
	"github.com/Nectar-Network/keeper-sdk/vault"
)

// Re-exported adapter types so SDK consumers can write keeper.ProtocolAdapter,
// keeper.Task, etc. without importing the adapters subpackage directly.
type (
	// ProtocolAdapter is implemented by every protocol integration.
	ProtocolAdapter = adapters.ProtocolAdapter
	// Task is one actionable unit of work an adapter discovers.
	Task = adapters.Task
	// Result is the outcome of executing a Task.
	Result = adapters.Result
	// VaultClient is the capital interface adapters use.
	VaultClient = adapters.VaultClient
)

// Keeper monitors protocols and executes profitable tasks using vault capital.
type Keeper struct {
	cfg      Config
	rpc      *soroban.Client
	kp       *keypair.Full
	vault    *vault.Client
	adapters []adapters.ProtocolAdapter
}

// NewKeeper creates a keeper from config. It parses the keeper secret and wires
// a vault client; it does not start polling (call Run).
func NewKeeper(cfg Config) (*Keeper, error) {
	kp, err := keypair.ParseFull(cfg.KeeperSecret)
	if err != nil {
		return nil, fmt.Errorf("parse keeper secret: %w", err)
	}
	rpc := soroban.NewClient(cfg.RpcURL)
	vc := vault.NewClient(rpc, kp, cfg.HorizonURL, cfg.Passphrase, cfg.VaultContract)
	return &Keeper{cfg: cfg, rpc: rpc, kp: kp, vault: vc}, nil
}

// AddAdapter registers a protocol adapter. Adapters are polled each cycle in the
// order added; tasks within a cycle run highest-priority first.
func (k *Keeper) AddAdapter(a adapters.ProtocolAdapter) { k.adapters = append(k.adapters, a) }

// RPC returns the shared Soroban client (useful when constructing adapters).
func (k *Keeper) RPC() *soroban.Client { return k.rpc }

// Keypair returns the keeper's signing keypair.
func (k *Keeper) Keypair() *keypair.Full { return k.kp }

// Config returns the keeper configuration.
func (k *Keeper) Config() Config { return k.cfg }

// Run starts the monitoring loop and blocks until the process exits. It is
// equivalent to RunContext(context.Background()) and returns an error
// immediately if no adapters are registered.
func (k *Keeper) Run() error { return k.RunContext(context.Background()) }

// RunContext starts the monitoring loop and blocks until ctx is cancelled,
// returning ctx.Err() for graceful shutdown (e.g. wire ctx to SIGINT/SIGTERM).
// It runs one cycle immediately, then once per PollInterval. Returns an error
// immediately if no adapters are registered.
func (k *Keeper) RunContext(ctx context.Context) error {
	if len(k.adapters) == 0 {
		return errors.New("keeper: no adapters registered")
	}
	slog.Info("keeper starting",
		"name", k.cfg.KeeperName, "adapters", len(k.adapters), "interval_s", k.cfg.PollInterval)
	ticker := time.NewTicker(time.Duration(k.cfg.PollInterval) * time.Second)
	defer ticker.Stop()
	for {
		k.cycle()
		select {
		case <-ctx.Done():
			slog.Info("keeper stopping", "reason", ctx.Err())
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// recoverStaleDraw makes the vault whole when a prior cycle left capital drawn
// but unreturned (e.g. a transient ReturnProceeds failure). Returns up to the
// outstanding draw from the keeper's USDC on hand — capped at the drawn amount,
// and a no-op on vaults deployed before get_keeper_draw existed.
func (k *Keeper) recoverStaleDraw() {
	if k.cfg.UsdcAddr == "" {
		return
	}
	drawn, err := vault.GetKeeperDraw(k.rpc, k.cfg.Passphrase, k.cfg.VaultContract, k.kp.Address())
	if err != nil || drawn <= 0 {
		return
	}
	usdc, err := dex.TokenBalance(k.rpc, k.cfg.Passphrase, k.cfg.UsdcAddr, k.kp.Address())
	if err != nil || usdc <= 0 {
		slog.Warn("outstanding vault draw but no USDC on hand — holding collateral", "drawn", drawn)
		return
	}
	ret := drawn
	if usdc < ret {
		ret = usdc
	}
	if err := k.vault.ReturnProceeds(ret, 0); err != nil {
		slog.Warn("stale-draw recovery failed", "drawn", drawn, "return", ret, "err", err)
		return
	}
	slog.Info("recovered stale vault draw", "drawn", drawn, "returned", ret)
}

// cycle runs every adapter once: scan tasks, sort by priority, execute.
func (k *Keeper) cycle() {
	k.recoverStaleDraw()
	for _, ad := range k.adapters {
		tasks, err := ad.GetTasks(k.rpc)
		if err != nil {
			slog.Warn("get tasks failed", "protocol", ad.Name(), "err", err)
			continue
		}
		adapters.SortByPriority(tasks)
		for _, task := range tasks {
			res, err := ad.Execute(k.rpc, k.kp, task, k.vault)
			if err != nil {
				slog.Warn("execute failed", "protocol", task.Protocol, "type", task.Type, "err", err)
				continue
			}
			switch {
			case res == nil:
				// nothing to record
			case res.Success:
				slog.Info("task executed", "protocol", task.Protocol, "type", task.Type,
					"proceeds", res.Proceeds, "profit", res.Profit, "tx", res.TxHash)
			case res.Note != "":
				slog.Info("task skipped", "protocol", task.Protocol, "note", res.Note)
			}
		}
	}
}
