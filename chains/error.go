package chains

import "errors"

var (
	ErrNotSupportChain = errors.New("not supported chain")
	ErrRLPDecode       = errors.New("rlp decode error")
)
