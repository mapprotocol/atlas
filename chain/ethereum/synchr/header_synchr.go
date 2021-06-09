package synchr

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"sync"
)

type latestBlockInfo struct {
	sync.Mutex
	fetchFinalizedDone bool
	height             uint64
}

type Synchr struct {
	descendantsUntilFinal uint64
	headerCache           HeaderCache
	headers               chan<- *types.Header
	loader                HeaderLoader
	log                   *logrus.Entry
	newHeaders            chan *types.Header
	oldHeaders            chan *types.Header
}

func NewSynchr(descendantsUntilFinal uint64, loader HeaderLoader, headers chan<- *types.Header, log *logrus.Entry) *Synchr {
	return &Synchr{
		descendantsUntilFinal: descendantsUntilFinal,
		headerCache:           *NewHeaderCache(descendantsUntilFinal + 1),
		headers:               headers,
		loader:                loader,
		log:                   log,
		newHeaders:            nil,
		oldHeaders:            nil,
	}
}

func (s *Synchr) StartSync(ctx context.Context, eg *errgroup.Group, initBlockHeight uint64) error {
	lbi := &latestBlockInfo{
		fetchFinalizedDone: false,
		height:             0,
	}
	s.newHeaders = make(chan *types.Header)
	s.oldHeaders = make(chan *types.Header)

	eg.Go(func() error {
		err := s.pollNewHeaders(ctx, lbi)
		close(s.newHeaders)
		return err
	})

	lbi.Lock()
	defer lbi.Unlock()
	latestHeader, err := s.loader.HeaderByNumber(ctx, nil)
	if err != nil {
		s.log.WithError(err).Error("Failed to retrieve latest header")
		close(s.headers)
		return err
	}
	if latestHeader.Number.Uint64() > lbi.height {
		lbi.height = latestHeader.Number.Uint64()
	}

	eg.Go(func() error {
		err := s.fetchFinalizedHeaders(ctx, initBlockHeight, lbi)
		close(s.oldHeaders)
		return err
	})

	eg.Go(func() error {
		for header := range s.oldHeaders {
			s.headers <- header
		}

		for header := range s.newHeaders {
			s.headers <- header
		}

		close(s.headers)
		return nil
	})

	return nil
}

func (s *Synchr) fetchFinalizedHeaders(ctx context.Context, initBlockHeight uint64, lbi *latestBlockInfo) error {
	return nil
}

func (s *Synchr) pollNewHeaders(ctx context.Context, lbi *latestBlockInfo) error {
	return nil
}
