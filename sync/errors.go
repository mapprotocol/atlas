package sync

import "errors"

var (
	ErrSyncInvalidInput  = errors.New("invalid input for sync")
	ErrExecutionReverted = errors.New("execution reverted")
)
