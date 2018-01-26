package events

import "errors"

var listenerExistError = errors.New("this event listener was exist")
var listenerNotExistError = errors.New("this event listener was not exist")

var eventExistError = errors.New("this event was exist")
var eventNotExistError = errors.New("this event was not exist")
