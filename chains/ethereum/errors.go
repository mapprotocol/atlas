package ethereum

import "errors"

var (
	errOlderBlockTime  = errors.New("timestamp older than parent")
	errUnknownAncestor = errors.New("unknown ancestor")
	errFutureBlock     = errors.New("block in the future")
	errInvalidNumber   = errors.New("invalid block number")
	errNotSupportChain = errors.New("not supported chain")
)
