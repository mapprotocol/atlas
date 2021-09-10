package ethconfig

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/mapprotocol/atlas/atlas/downloader"
	"github.com/mapprotocol/atlas/atlas/gasprice"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	"github.com/mapprotocol/atlas/core/chain"
	"github.com/mapprotocol/atlas/core/txsdetails"
	"github.com/mapprotocol/atlas/miner"
	params2 "github.com/mapprotocol/atlas/params"
	"math/big"
	"time"
)

// FullNodeGPO contains default gasprice oracle settings for full node.
var FullNodeGPO = gasprice.Config{
	Blocks:     20,
	Percentile: 60,
	MaxPrice:   gasprice.DefaultMaxPrice,
}

// LightClientGPO contains default gasprice oracle settings for light client.
var LightClientGPO = gasprice.Config{
	Blocks:     2,
	Percentile: 60,
	MaxPrice:   gasprice.DefaultMaxPrice,
}

// Defaults contains default settings for use on the Ethereum main net.
var Defaults = Config{
	SyncMode: downloader.FastSync,
	NetworkId:               params2.MainnetNetWorkID,
	TxLookupLimit:           2350000,
	LightPeers:              100,
	LightServ:               0,
	UltraLightFraction:      75,
	DatabaseCache:           512,
	TrieCleanCache:          154,
	TrieCleanCacheJournal:   "triecache",
	TrieCleanCacheRejournal: 60 * time.Minute,
	TrieDirtyCache:          256,
	TrieTimeout:             60 * time.Minute,
	SnapshotCache:           102,
	GatewayFee:              big.NewInt(0),
	Miner: miner.Config{
		GasFloor: 8000000,
		GasCeil:  8000000,
		GasPrice: big.NewInt(params.GWei),
		Recommit: 3 * time.Second,
	},
	TxPool:      txsdetails.DefaultTxPoolConfig,
	RPCGasCap:   25000000,
	GPO:         FullNodeGPO,
	RPCTxFeeCap: 1, // 1 ether
	Istanbul: *istanbul.DefaultConfig,
}

//func init() {
//	home := os.Getenv("HOME")
//	if home == "" {
//		if user, err := user.Current(); err == nil {
//			home = user.HomeDir
//		}
//	}
//	if runtime.GOOS == "darwin" {
//		Defaults.Ethash.DatasetDir = filepath.Join(home, "Library", "Ethash")
//	} else if runtime.GOOS == "windows" {
//		localappdata := os.Getenv("LOCALAPPDATA")
//		if localappdata != "" {
//			Defaults.Ethash.DatasetDir = filepath.Join(localappdata, "Ethash")
//		} else {
//			Defaults.Ethash.DatasetDir = filepath.Join(home, "AppData", "Local", "Ethash")
//		}
//	} else {
//		Defaults.Ethash.DatasetDir = filepath.Join(home, ".ethash")
//	}
//}

//go:generate gencodec -type Config -formats toml -out gen_config.go

type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Ethereum main net block is used.
	Genesis *chain.Genesis `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	SyncMode  downloader.SyncMode

	// This can be set to list of enrtree:// URLs which will be queried for
	// for nodes to connect to.
	DiscoveryURLs []string

	NoPruning  bool // Whether to disable pruning and flush everything to disk
	NoPrefetch bool // Whether to disable prefetching and only load state on demand

	TxLookupLimit uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.

	// Whitelist of required block number -> hash values to accept
	Whitelist map[uint64]common.Hash `toml:"-"`

	// Light client options
	LightServ    int  `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	LightIngress int  `toml:",omitempty"` // Incoming bandwidth limit for light servers
	LightEgress  int  `toml:",omitempty"` // Outgoing bandwidth limit for light servers
	LightPeers   int  `toml:",omitempty"` // Maximum number of LES client peers
	LightNoPrune bool `toml:",omitempty"` // Whether to disable light chain pruning
	// Minimum gateway fee value to serve a transaction from a light client
	GatewayFee *big.Int `toml:",omitempty"`
	// Validator is the address used to sign consensus messages. Also the address for block transaction rewards.
	Validator common.Address `toml:",omitempty"`
	// TxFeeRecipient is the GatewayFeeRecipient light clients need to specify in order for their transactions to be accepted by this node.
	TxFeeRecipient common.Address `toml:",omitempty"`
	BLSbase        common.Address `toml:",omitempty"`

	// Ultra Light client options
	UltraLightServers      []string `toml:",omitempty"` // List of trusted ultra light servers
	UltraLightFraction     int      `toml:",omitempty"` // Percentage of trusted servers to accept an announcement
	UltraLightOnlyAnnounce bool     `toml:",omitempty"` // Whether to only announce headers, or also serve them

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	DatabaseFreezer    string

	TrieCleanCache          int
	TrieCleanCacheJournal   string        `toml:",omitempty"` // Disk journal directory for trie cache to survive node restarts
	TrieCleanCacheRejournal time.Duration `toml:",omitempty"` // Time interval to regenerate the journal for clean cache
	TrieDirtyCache          int
	TrieTimeout             time.Duration
	SnapshotCache           int

	// Mining options
	Miner miner.Config

	// Transaction pool options
	TxPool txsdetails.TxPoolConfig

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Istanbul options
	Istanbul istanbul.Config

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// Type of the EWASM interpreter ("" for default)
	EWASMInterpreter string

	// Type of the EVM interpreter ("" for default)
	EVMInterpreter string

	// RPCGasCap is the global gas cap for eth-call variants.
	RPCGasCap uint64 `toml:",omitempty"`

	// RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for
	// send-transction variants. The unit is ether.
	RPCTxFeeCap float64 `toml:",omitempty"`

	// Checkpoint is a hardcoded checkpoint which can be nil.
	Checkpoint *params.TrustedCheckpoint `toml:",omitempty"`

	// CheckpointOracle is the configuration for checkpoint oracle.
	CheckpointOracle *params.CheckpointOracleConfig `toml:",omitempty"`

	// Churrito block override (TODO: remove after the fork)
	OverrideChurrito *big.Int `toml:",omitempty"`

	// Donut block override (TODO: remove after the fork)
	OverrideDonut *big.Int `toml:",omitempty"`

	// Gas Price Oracle options
	GPO gasprice.Config
}