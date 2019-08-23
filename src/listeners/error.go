package listeners

import "errors"

var roomNotExistError = errors.New("live room was not exist")
var listenerExistError = errors.New("this live has a listener")
var listenerNotExistError = errors.New("this live has not a listener")
