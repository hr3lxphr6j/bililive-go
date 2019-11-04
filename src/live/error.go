package live

import (
	"errors"
)

var (
	ErrRoomNotExist     = errors.New("room not exists")
	ErrRoomUrlIncorrect = errors.New("room url incorrect")
	ErrInternalError    = errors.New("internal error")
)
