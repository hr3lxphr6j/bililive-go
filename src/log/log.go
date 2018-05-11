package log

import (
	"context"
	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
	"github.com/sirupsen/logrus"
	"os"
)

func info2Fields(live api.Live) logrus.Fields {
	return logrus.Fields{
		"Url":      live.GetRawUrl(),
		"HostName": live.GetCachedInfo().HostName,
		"RoomName": live.GetCachedInfo().RoomName,
	}
}

func registerEventLog(ed events.IEventDispatcher, logger *interfaces.Logger) {

	targetEvents := []events.EventType{
		listeners.ListenStart,
		listeners.LiveStart,
		listeners.LiveEnd,
		recorders.RecordeStart,
		recorders.RecordeStop,
	}

	for _, e := range targetEvents {
		ed.AddEventListener(e, events.NewEventListener(func(event *events.Event) {
			logger.WithFields(info2Fields(event.Object.(api.Live))).Info(event.Type)
		}))
	}

}

func NewLogger(ctx context.Context) *interfaces.Logger {
	inst := instance.GetInstance(ctx)

	logLevel := logrus.InfoLevel
	switch inst.Config.LogLevel {
	case "panic":
		logLevel = logrus.PanicLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "debug":
		logLevel = logrus.DebugLevel
	default:
		logLevel = logrus.InfoLevel
	}

	logger := &interfaces.Logger{Logger: &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
		Hooks: make(logrus.LevelHooks),
		Level: logLevel,
	}}

	inst.Logger = logger

	registerEventLog(inst.EventDispatcher.(events.IEventDispatcher), logger)

	return logger
}
