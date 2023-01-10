package define

import "errors"

var (
	GetIndexError          = errors.New("get Index nil(no Address)")
	NoTargetValidatorError = errors.New("not find target validator")
	BigSubValue            = errors.New("not enough map")
)
