package recorders

import "errors"

var recorderExistError = errors.New("this recorder is exist")
var recorderNotExistError = errors.New("this recorder is not exist")
