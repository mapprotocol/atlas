// Copyright 2021 MAP Protocol Authors.
// This file is part of MAP Protocol.

// MAP Protocol is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// MAP Protocol is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with MAP Protocol.  If not, see <http://www.gnu.org/licenses/>.

// Contains the atlas command usage template and generator.

package cmd

import (
	"io"
	"sort"

	"github.com/mapprotocol/atlas/cmd/utils"
	"github.com/mapprotocol/atlas/helper/debug"
	"github.com/mapprotocol/atlas/helper/flags"
	"gopkg.in/urfave/cli.v1"
)

// AppHelpFlagGroups is the application flags, grouped by functionality.
var AppHelpFlagGroups = []flags.FlagGroup{
	{
		Name: "MAPPROTOCOL",
		Flags: []cli.Flag{
			configFileFlag,
			utils.DataDirFlag,
			utils.AncientFlag,
			utils.MinFreeDiskSpaceFlag,
			utils.KeyStoreDirFlag,
			utils.USBFlag,
			utils.SmartCardDaemonPathFlag,
			utils.NetworkIdFlag,
			utils.MainnetFlag,
			utils.TestnetFlag,
			utils.SyncModeFlag,
			utils.ExitWhenSyncedFlag,
			utils.GCModeFlag,
			utils.TxLookupLimitFlag,
			utils.EthStatsURLFlag,
			utils.IdentityFlag,
			utils.LightKDFFlag,
			utils.WhitelistFlag,
			utils.VerifyCheckPointFlag,
		},
	},
	{
		Name: "LIGHT CLIENT",
		Flags: []cli.Flag{
			utils.LightServeFlag,
			utils.LightIngressFlag,
			utils.LightEgressFlag,
			utils.LightMaxPeersFlag,
			utils.UltraLightServersFlag,
			utils.UltraLightFractionFlag,
			utils.UltraLightOnlyAnnounceFlag,
			utils.LightNoPruneFlag,
			utils.LightNoSyncServeFlag,
		},
	},
	{
		Name: "DEVELOPER CHAIN",
		Flags: []cli.Flag{
			utils.DeveloperFlag,
		},
	},
	{
		Name: "SINGLE CHAIN",
		Flags: []cli.Flag{
			utils.SingleFlag,
		},
	},
	{
		Name: "TRANSACTION POOL",
		Flags: []cli.Flag{
			utils.TxPoolLocalsFlag,
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
		},
	},
	{
		Name: "PERFORMANCE TUNING",
		Flags: []cli.Flag{
			utils.CacheFlag,
			utils.CacheDatabaseFlag,
			utils.CacheTrieFlag,
			utils.CacheTrieJournalFlag,
			utils.CacheTrieRejournalFlag,
			utils.CacheGCFlag,
			utils.CacheSnapshotFlag,
			utils.CacheNoPrefetchFlag,
			utils.CachePreimagesFlag,
		},
	},
	{
		Name: "ACCOUNT",
		Flags: []cli.Flag{
			utils.UnlockedAccountFlag,
			utils.PasswordFileFlag,
			utils.ExternalSignerFlag,
			utils.InsecureUnlockAllowedFlag,
		},
	},
	{
		Name: "API AND CONSOLE",
		Flags: []cli.Flag{
			utils.IPCDisabledFlag,
			utils.IPCPathFlag,
			utils.HTTPEnabledFlag,
			utils.HTTPListenAddrFlag,
			utils.HTTPPortFlag,
			utils.HTTPApiFlag,
			utils.HTTPPathPrefixFlag,
			utils.HTTPCORSDomainFlag,
			utils.HTTPVirtualHostsFlag,
			utils.WSEnabledFlag,
			utils.WSListenAddrFlag,
			utils.WSPortFlag,
			utils.WSApiFlag,
			utils.WSPathPrefixFlag,
			utils.WSAllowedOriginsFlag,
			utils.GraphQLEnabledFlag,
			utils.GraphQLCORSDomainFlag,
			utils.GraphQLVirtualHostsFlag,
			utils.RPCGlobalGasCapFlag,
			utils.RPCGlobalTxFeeCapFlag,
			utils.AllowUnprotectedTxs,
			utils.JSpathFlag,
			utils.ExecFlag,
			utils.PreloadJSFlag,
		},
	},
	{
		Name: "NETWORKING",
		Flags: []cli.Flag{
			utils.BootnodesFlag,
			utils.DNSDiscoveryFlag,
			utils.ListenPortFlag,
			utils.MaxPeersFlag,
			utils.MaxPendingPeersFlag,
			utils.NATFlag,
			utils.NoDiscoverFlag,
			utils.DiscoveryV5Flag,
			utils.NetrestrictFlag,
			utils.NodeKeyFileFlag,
			utils.NodeKeyHexFlag,
		},
	},
	{
		Name: "MINER",
		Flags: []cli.Flag{
			utils.MiningEnabledFlag,
			utils.MinerThreadsFlag,
			//utils.MinerNotifyFlag,
			//utils.MinerNotifyFullFlag,
			utils.MinerGasPriceFlag,
			//utils.MinerGasLimitFlag,
			utils.MinerValidatorFlag,
			utils.MinerExtraDataFlag,
			//utils.MinerRecommitIntervalFlag,
			//utils.MinerNoVerifyFlag,
		},
	},
	{
		Name: "GAS PRICE ORACLE",
		Flags: []cli.Flag{
			utils.GpoBlocksFlag,
			utils.GpoPercentileFlag,
			utils.GpoMaxGasPriceFlag,
		},
	},
	{
		Name: "VIRTUAL MACHINE",
		Flags: []cli.Flag{
			utils.VMEnableDebugFlag,
		},
	},
	{
		Name: "LOGGING AND DEBUGGING",
		Flags: append([]cli.Flag{
			utils.FakePoWFlag,
			utils.NoCompactionFlag,
		}, debug.Flags...),
	},
	{
		Name:  "METRICS AND STATS",
		Flags: metricsFlags,
	},
	{
		Name: "MISC",
		Flags: []cli.Flag{
			utils.SnapshotFlag,
			utils.BloomFilterSizeFlag,
			cli.HelpFlag,
			utils.CatalystFlag,
		},
	},
}

func init() {
	// Override the default app help template
	cli.AppHelpTemplate = flags.AppHelpTemplate

	// Override the default app help printer, but only for the global app help
	originalHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, tmpl string, data interface{}) {
		if tmpl == flags.AppHelpTemplate {
			// Iterate over all the flags and add any uncategorized ones
			categorized := make(map[string]struct{})
			for _, group := range AppHelpFlagGroups {
				for _, flag := range group.Flags {
					categorized[flag.String()] = struct{}{}
				}
			}
			deprecated := make(map[string]struct{})
			for _, flag := range utils.DeprecatedFlags {
				deprecated[flag.String()] = struct{}{}
			}
			// Only add uncategorized flags if they are not deprecated
			var uncategorized []cli.Flag
			for _, flag := range data.(*cli.App).Flags {
				if _, ok := categorized[flag.String()]; !ok {
					if _, ok := deprecated[flag.String()]; !ok {
						uncategorized = append(uncategorized, flag)
					}
				}
			}
			if len(uncategorized) > 0 {
				// Append all ungategorized options to the misc group
				miscs := len(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags)
				AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = append(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags, uncategorized...)

				// Make sure they are removed afterwards
				defer func() {
					AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags[:miscs]
				}()
			}
			// Render out custom usage screen
			originalHelpPrinter(w, tmpl, flags.HelpData{App: data, FlagGroups: AppHelpFlagGroups})
		} else if tmpl == flags.CommandHelpTemplate {
			// Iterate over all command specific flags and categorize them
			categorized := make(map[string][]cli.Flag)
			for _, flag := range data.(cli.Command).Flags {
				if _, ok := categorized[flag.String()]; !ok {
					categorized[flags.FlagCategory(flag, AppHelpFlagGroups)] = append(categorized[flags.FlagCategory(flag, AppHelpFlagGroups)], flag)
				}
			}

			// sort to get a stable ordering
			sorted := make([]flags.FlagGroup, 0, len(categorized))
			for cat, flgs := range categorized {
				sorted = append(sorted, flags.FlagGroup{Name: cat, Flags: flgs})
			}
			sort.Sort(flags.ByCategory(sorted))

			// add sorted array to data and render with default printer
			originalHelpPrinter(w, tmpl, map[string]interface{}{
				"cmd":              data,
				"categorizedFlags": sorted,
			})
		} else {
			originalHelpPrinter(w, tmpl, data)
		}
	}
}
