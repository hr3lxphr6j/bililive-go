package zaplogger

import (
	"errors"
	"fmt"
)

var (
	ErrFailedAppend = errors.New("failed append")
)

func FailedAppendError(err error) error {
	return fmt.Errorf("%w: %v", ErrFailedAppend, err)
}
