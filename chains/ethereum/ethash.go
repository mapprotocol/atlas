package ethereum

import (
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"path/filepath"
)

func NewEthash(dir string) *ethash.Ethash {
	DatasetDir := filepath.Join(dir, "Ethash")
	return ethash.New(ethash.Config{
		CacheDir:         DatasetDir,
		CachesInMem:      2,
		CachesOnDisk:     3,
		CachesLockMmap:   false,
		DatasetDir:       DatasetDir,
		DatasetsInMem:    1,
		DatasetsOnDisk:   2,
		DatasetsLockMmap: false,
		PowMode:          ethash.ModeNormal,
	}, nil, false)
}
