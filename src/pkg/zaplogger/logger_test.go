package zaplogger

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	lg := GetLogger()
	str := `{"Hello": "INNER"}`
	lg.Info("failed", zap.String("hello", str))

	lg1 := lg.With(zap.String("pkg", "test"))
	lg1.Info("failed", zap.String("hello", str))

	lg2 := lg.With(zap.String("traceId", "test2"))
	lg2.Info("failed", zap.String("hello", str))

	lg2.Sugar().With("topic", "topicname").Error("failed suger")

	// lg = lg.With(zap.String("pkg", "test"))
	// lg.Info("using with",
	// 	zap.Duration("time", time.Second),
	// 	zap.Int("int", 16))
}
