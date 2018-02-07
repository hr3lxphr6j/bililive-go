package recorders

import "errors"

var recorderExistError = errors.New("this recorder has a listener")
var recorderNotExistError = errors.New("this recorder has not a listener")
