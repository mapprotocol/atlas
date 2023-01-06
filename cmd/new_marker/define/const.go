package define

import "errors"

var (
	GetIndexError          = errors.New("get Index nil(no Address)")
	NoTargetValidatorError = errors.New("not find target validator")
	bigSubValue            = errors.New("not enough map")
	isContinueError        = true
)
