// Copyright 2017 The go-tau Authors
// This file is part of go-tau.
//
// go-tau is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-tau is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-tau. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"reflect"
	"unicode"

	"gopkg.in/urfave/cli.v1"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/cmd/utils"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/node"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/params"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau"
	"github.com/naoina/toml"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, rpcFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type taustatsConfig struct {
	URL string `toml:",omitempty"`
}

type gtauConfig struct {
	Tau      tau.Config
	Node     node.Config
	Taustats taustatsConfig
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "tau", "shh")
	cfg.WSModules = append(cfg.WSModules, "tau", "shh")
	cfg.IPCPath = "gtau.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gtauConfig) {
	// Load defaults.
	cfg := gtauConfig{
		Tau:  tau.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	// Apply flags.
	// Node
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	// Tau
	utils.SetTauConfig(ctx, stack, &cfg.Tau)
	// Tau status service
	if ctx.GlobalIsSet(utils.TauStatsURLFlag.Name) {
		cfg.Taustats.URL = ctx.GlobalString(utils.TauStatsURLFlag.Name)
	}

	return stack, cfg
}

func makeFullNode(ctx *cli.Context) *node.Node {

	// Configure 
	stack, cfg := makeConfigNode(ctx)

	// Register tauservice
	utils.RegisterTauService(stack, &cfg.Tau)

	// Add the tau status daemon if requested.
	if cfg.Taustats.URL != "" {
		utils.RegisterTauStatsService(stack, cfg.Taustats.URL)
	}

	return stack
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.Tau.Genesis != nil {
		cfg.Tau.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}
