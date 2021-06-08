package synchr

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"sync"
)

type latestBlockInfo struct {
	sync.Mutex
	fetchFinalizedDone bool
	height             uint64
}

type synchr struct {
	descendantsUntilFinal uint64
	headerCache           HeaderCache
	headers               chan<- *types.Header
	loader                HeaderLoader
	log                   *logrus.Entry
	newHeaders            chan *types.Header
	oldHeaders            chan *types.Header
}

func NewSynchr(descendantsUntilFinal uint64, loader HeaderLoader, headers chan<- *types.Header, log *logrus.Entry) *synchr {
	return &synchr{
		descendantsUntilFinal: descendantsUntilFinal,
		headerCache:           *NewHeaderCache(descendantsUntilFinal + 1),
		headers:               headers,
		loader:                loader,
		log:                   log,
		newHeaders:            nil,
		oldHeaders:            nil,
	}
}
