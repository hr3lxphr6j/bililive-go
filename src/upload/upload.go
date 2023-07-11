package upload

import (
	"github.com/matyle/bililive-go/src/pkg/zaplogger"
	"go.uber.org/zap"
)

var (
	log = zaplogger.GetLogger().With(zap.String("pkg", "upload"))
)

type Uploader interface {
	Upload() error
}
