package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alecsavvy/mojave/app"
	"github.com/alecsavvy/mojave/config"
	"github.com/urfave/cli/v3"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

func main() {
	root := &cli.Command{
		Name:  "mojave",
		Usage: "mojave node and testnet",
		Commands: []*cli.Command{
			runCmd,
			testnetCmd,
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := root.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

var runCmd = &cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "run the mojave node",
	Action: func(ctx context.Context, c *cli.Command) error {
		homeDir := os.TempDir() + "/mojave-dev-" + time.Now().Format("20060102150405")
		cmtConfig := cfg.DefaultConfig()
		cmtConfig.SetRoot(homeDir)

		if _, _, _, err := config.InitFilesWithConfig(cmtConfig); err != nil {
			return err
		}

		a := app.NewApp(cmtConfig)
		return a.Run(ctx)
	},
}

var testnetCmd = &cli.Command{
	Name:  "testnet",
	Usage: "run a local multi-validator testnet",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "validators",
			Aliases: []string{"n"},
			Value:   4,
			Usage:   "number of validators",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		n := c.Int("validators")
		if n < 1 {
			return fmt.Errorf("validators must be at least 1")
		}

		testnetDir := filepath.Join(os.TempDir(), "mojave-testnet-"+time.Now().Format("20060102150405"))
		if err := os.MkdirAll(testnetDir, 0755); err != nil {
			return err
		}
		log.Printf("Testnet root: %s", testnetDir)

		// 1) Build configs (no persistent_peers yet) and generate keys to get node IDs
		type nodeInfo struct {
			config *cfg.Config
			nodeID string
			pubKey types.GenesisValidator
		}
		nodes := make([]nodeInfo, n)

		for i := 0; i < n; i++ {
			nodeDir := filepath.Join(testnetDir, "node"+strconv.Itoa(i))
			cmtConfig := cfg.DefaultConfig()
			cmtConfig.SetRoot(nodeDir)
			p2pPort := 26656 + i*4
			rpcPort := 26657 + i*4
			cmtConfig.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", p2pPort)
			cmtConfig.RPC.ListenAddress = fmt.Sprintf("tcp://127.0.0.1:%d", rpcPort)
			cmtConfig.P2P.AllowDuplicateIP = true

			_, nodeKey, pubKey, err := config.GenValidatorKeys(cmtConfig)
			if err != nil {
				return fmt.Errorf("node %d: %w", i, err)
			}

			nodes[i] = nodeInfo{
				config: cmtConfig,
				nodeID: string(nodeKey.ID()),
				pubKey: types.GenesisValidator{Address: pubKey.Address(), PubKey: pubKey, Power: 10},
			}
		}

		// 2) Set persistent_peers on each config
		for i := 0; i < n; i++ {
			var peers []string
			for j := 0; j < n; j++ {
				if j == i {
					continue
				}
				p2pPort := 26656 + j*4
				peers = append(peers, nodes[j].nodeID+"@127.0.0.1:"+strconv.Itoa(p2pPort))
			}
			nodes[i].config.P2P.PersistentPeers = strings.Join(peers, ",")
		}

		// 3) Write configs and shared genesis
		chainID := "mojave-testnet-" + time.Now().Format("20060102150405")
		genDoc := &types.GenesisDoc{
			ChainID:         chainID,
			GenesisTime:     cmttime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
			Validators:      make([]types.GenesisValidator, n),
		}
		for i := 0; i < n; i++ {
			genDoc.Validators[i] = nodes[i].pubKey
		}

		for i := 0; i < n; i++ {
			cfg.WriteConfigFile(filepath.Join(nodes[i].config.RootDir, "config", "config.toml"), nodes[i].config)
			if err := genDoc.SaveAs(nodes[i].config.GenesisFile()); err != nil {
				return fmt.Errorf("node %d write genesis: %w", i, err)
			}
		}

		// 4) Run all nodes
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			nodeConfig := nodes[i].config
			go func() {
				defer wg.Done()
				a := app.NewApp(nodeConfig)
				_ = a.Run(ctx)
			}()
		}
		wg.Wait()
		return nil
	},
}
