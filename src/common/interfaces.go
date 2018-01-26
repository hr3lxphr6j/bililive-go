package common

type Runnable interface {
	Start() error
	Close()
}