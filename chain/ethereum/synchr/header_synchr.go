package synchr

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"math/big"
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
	syncedUpUntil := initBlockHeight

	for {
		lbi.Lock()
		latestFinalizedHeight := saturatingSub(lbi.height, s.descendantsUntilFinal)
		if syncedUpUntil >= latestFinalizedHeight {
			// Signals to pollNewHeaders that new headers can be forwarded now
			lbi.fetchFinalizedDone = true
			lbi.Unlock()

			s.log.WithField("blockNumber", syncedUpUntil).Debug("Done retrieving finalized headers")

			break
		}
		lbi.Unlock()

		header, err := s.loader.HeaderByNumber(ctx, new(big.Int).SetUint64(syncedUpUntil+1))
		if err != nil {
			s.log.WithField(
				"blockNumber", syncedUpUntil+1,
			).WithError(err).Error("Failed to retrieve finalized header")
			return err
		}

		s.log.WithFields(logrus.Fields{
			"blockHash":   header.Hash().Hex(),
			"blockNumber": syncedUpUntil + 1,
		}).Debug("Retrieved finalized header")

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.oldHeaders <- header
		}
		syncedUpUntil++
	}

	return nil
}

func (s *Synchr) pollNewHeaders(ctx context.Context, lbi *latestBlockInfo) error {
	headers := make(chan *types.Header)
	var headersSubscriptionErr <-chan error

	subscription, err := s.loader.SubscribeNewHead(ctx, headers)
	if err != nil {
		s.log.WithError(err).Error("Failed to subscribe to new headers")
		return err
	}
	headersSubscriptionErr = subscription.Err()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-headersSubscriptionErr:
			return err
		case header := <-headers:
			s.headerCache.Add(header)
			lbi.Lock()
			lbi.height = header.Number.Uint64()

			s.log.WithFields(logrus.Fields{
				"blockHash":   header.Hash().Hex(),
				"blockNumber": lbi.height,
			}).Debug("Witnessed new header")

			if lbi.fetchFinalizedDone {
				err = s.forwardAncestry(ctx, header.Hash(), saturatingSub(lbi.height, s.descendantsUntilFinal))
				if err != nil {
					s.log.WithFields(logrus.Fields{
						"blockHash":   header.Hash().Hex(),
						"blockNumber": lbi.height,
					}).WithError(err).Error("Failed to forward header and its ancestors")
				}
			}
			lbi.Unlock()
		}
	}
}

func (s *Synchr) forwardAncestry(ctx context.Context, hash common.Hash, oldestHeight uint64) error {
	item, exists := s.headerCache.Get(hash)
	if !exists {
		header, err := s.loader.HeaderByHash(ctx, hash)
		if err != nil {
			return err
		}

		// If a header is too old, it cannot be inserted. We can assume it's already been forwarded
		if !s.headerCache.Add(header) {
			return nil
		}
		item, _ = s.headerCache.Get(hash)
	}

	if item.Forwarded {
		return nil
	}

	if item.Header.Number.Uint64() > oldestHeight {
		err := s.forwardAncestry(ctx, item.Header.ParentHash, oldestHeight)
		if err != nil {
			return err
		}
	}

	s.newHeaders <- item.Header
	item.Forwarded = true
	return nil
}

// Subtraction but returns 0 when r > l
func saturatingSub(l uint64, r uint64) uint64 {
	if r > l {
		return 0
	}
	return l - r
}
