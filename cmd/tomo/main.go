// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/tomochain/tomochain/accounts"
	"github.com/tomochain/tomochain/accounts/keystore"
	"github.com/tomochain/tomochain/cmd/utils"
	"github.com/tomochain/tomochain/consensus/posv"
	"github.com/tomochain/tomochain/console"
	"github.com/tomochain/tomochain/core"
	"github.com/tomochain/tomochain/eth"
	"github.com/tomochain/tomochain/ethclient"
	"github.com/tomochain/tomochain/internal/debug"
	"github.com/tomochain/tomochain/log"
	"github.com/tomochain/tomochain/metrics"
	"github.com/tomochain/tomochain/node"
	"gopkg.in/urfave/cli.v1"
)

const (
	clientIdentifier = "tomo" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, "the tomochain command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		//utils.NoUSBFlag,
		//utils.EthashCacheDirFlag,
		//utils.EthashCachesInMemoryFlag,
		//utils.EthashCachesOnDiskFlag,
		//utils.EthashDatasetDirFlag,
		//utils.EthashDatasetsInMemoryFlag,
		//utils.EthashDatasetsOnDiskFlag,
		utils.TomoXEnabledFlag,
		utils.TomoXDataDirFlag,
		utils.TomoXDBEngineFlag,
		utils.TomoXDBConnectionUrlFlag,
		utils.TomoXDBReplicaSetNameFlag,
		utils.TomoXDBNameFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.FastSyncFlag,
		utils.LightModeFlag,
		utils.SyncModeFlag,
		utils.GCModeFlag,
		//utils.LightServFlag,
		//utils.LightPeersFlag,
		//utils.LightKDFFlag,
		//utils.CacheFlag,
		//utils.CacheDatabaseFlag,
		//utils.CacheGCFlag,
		//utils.TrieCacheGenFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.EtherbaseFlag,
		utils.GasPriceFlag,
		utils.StakerThreadsFlag,
		utils.StakingEnabledFlag,
		utils.TargetGasLimitFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		//utils.DiscoveryV5Flag,
		//utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		//utils.DeveloperFlag,
		//utils.DeveloperPeriodFlag,
		//utils.TestnetFlag,
		//utils.RinkebyFlag,
		//utils.VMEnableDebugFlag,
		utils.TomoTestnetFlag,
		utils.RewoundFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.EthStatsURLFlag,
		//utils.FakePoWFlag,
		//utils.NoCompactionFlag,
		//utils.GpoBlocksFlag,
		//utils.GpoPercentileFlag,
		//utils.ExtraDataFlag,
		configFileFlag,
		utils.AnnounceTxsFlag,
		utils.StoreRewardFlag,
		utils.RollbackFlag,
		utils.ReexecFlag,
	}

	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCApiFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
	}

	whisperFlags = []cli.Flag{
		utils.WhisperEnabledFlag,
		utils.WhisperMaxMessageSizeFlag,
		utils.WhisperMinPOWFlag,
	}
	metricsFlags = []cli.Flag{
		utils.MetricsEnabledFlag,
		utils.MetricsEnabledExpensiveFlag,
		utils.MetricsHTTPFlag,
		utils.MetricsPortFlag,
	}
)

func init() {
	// Initialize the CLI app and start tomo
	app.Action = tomo
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright (c) 2018 Tomochain"
	app.Commands = []cli.Command{
		// See chaincmd.go:
		initCommand,
		importCommand,
		exportCommand,
		removedbCommand,
		dumpCommand,
		// See accountcmd.go:
		accountCommand,
		walletCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		javascriptCommand,
		// See dbcmd.go:
		dbCommand,
		// See misccmd.go:
		versionCommand,
		// See config.go
		dumpConfigCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, whisperFlags...)
	app.Flags = append(app.Flags, metricsFlags...)

	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		if err := debug.Setup(ctx); err != nil {
			return err
		}

		// Start metrics export if enabled
		utils.SetupMetrics(ctx)

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(3 * time.Second)

		utils.SetupNetwork(ctx)
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// tomo is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func tomo(ctx *cli.Context) error {
	node, cfg := makeFullNode(ctx)
	startNode(ctx, node, cfg)
	node.Wait()
	return nil
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *node.Node, cfg tomoConfig) {
	// Start up the node itself
	utils.StartNode(stack)

	// Unlock any account specifically requested
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	if ctx.GlobalIsSet(utils.UnlockedAccountFlag.Name) {
		cfg.Account.Unlocks = strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	}

	if ctx.GlobalIsSet(utils.PasswordFileFlag.Name) {
		cfg.Account.Passwords = utils.MakePasswordList(ctx)
	}

	for i, account := range cfg.Account.Unlocks {
		if trimmed := strings.TrimSpace(account); trimmed != "" {
			unlockAccount(ctx, ks, trimmed, i, cfg.Account.Passwords)
		}
	}
	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	go func() {
		// Create an chain state reader for self-derivation
		rpcClient, err := stack.Attach()
		if err != nil {
			utils.Fatalf("Failed to attach to self: %v", err)
		}
		stateReader := ethclient.NewClient(rpcClient)

		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				if event.Wallet.URL().Scheme == "ledger" {
					event.Wallet.SelfDerive(accounts.DefaultLedgerBaseDerivationPath, stateReader)
				} else {
					event.Wallet.SelfDerive(accounts.DefaultBaseDerivationPath, stateReader)
				}

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()
	// Start auxiliary services if enabled

	// Mining only makes sense if a full Ethereum node is running
	if ctx.GlobalBool(utils.LightModeFlag.Name) || ctx.GlobalString(utils.SyncModeFlag.Name) == "light" {
		utils.Fatalf("Light clients do not support staking")
	}
	var ethereum *eth.Ethereum
	if err := stack.Service(&ethereum); err != nil {
		utils.Fatalf("Ethereum service not running: %v", err)
	}
	if _, ok := ethereum.Engine().(*posv.Posv); ok {
		go func() {
			started := false
			ok, err := ethereum.ValidateMasternode()
			if err != nil {
				log.Warn("Cannot get etherbase", "err", err)
			}
			if ok {
				log.Info("Masternode found. Enabling staking mode...")
				if threads := ctx.GlobalInt(utils.StakerThreadsFlag.Name); threads > 0 {
					if th, ok := ethereum.Engine().(concurrentEngine); ok {
						th.SetThreads(threads)
					}
				}
				// Set the gas price to the limits from the CLI and start mining
				ethereum.TxPool().SetGasPrice(cfg.Eth.GasPrice)
				if err := ethereum.StartStaking(true); err != nil {
					utils.Fatalf("Failed to start staking: %v", err)
				}
				started = true
				log.Info("Enabled mining node!!!")
			}
			defer close(core.CheckpointCh)
			for range core.CheckpointCh {
				log.Info("Checkpoint!!! It's time to reconcile node's state...")
				ok, err := ethereum.ValidateMasternode()
				if err != nil {
					log.Warn("Cannot get etherbase", "err", err)
				}
				if !ok {
					if started {
						log.Info("Only masternode can propose and verify blocks. Cancelling staking on this node...")
						ethereum.StopStaking()
						started = false
						log.Info("Cancelled mining mode!!!")
					}
				} else if !started {
					log.Info("Masternode found. Enabling staking mode...")
					if threads := ctx.GlobalInt(utils.StakerThreadsFlag.Name); threads > 0 {
						if th, ok := ethereum.Engine().(concurrentEngine); ok {
							th.SetThreads(threads)
						}
					}
					// Set the gas price to the limits from the CLI and start mining
					ethereum.TxPool().SetGasPrice(cfg.Eth.GasPrice)
					if err := ethereum.StartStaking(true); err != nil {
						utils.Fatalf("Failed to start staking: %v", err)
					}
					started = true
					log.Info("Enabled mining node!!!")
				}
			}
		}()
	}
}

// Interface to check a consensus engine can support many threads.
type concurrentEngine interface {
	SetThreads(threads int)
}
