package blog

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var l *zap.Logger

var Debugf = func(ctx context.Context, format string, args ...interface{}) {}
var Infof = func(ctx context.Context, format string, args ...interface{}) {}
var Warnf = func(ctx context.Context, format string, args ...interface{}) {}
var Errorf = func(ctx context.Context, format string, args ...interface{}) {}
var Criticalf = func(ctx context.Context, format string, args ...interface{}) {}

var Debug = func(ctx context.Context, args ...interface{}) {}
var Info = func(ctx context.Context, args ...interface{}) {}
var Warn = func(ctx context.Context, args ...interface{}) {}
var Error = func(ctx context.Context, args ...interface{}) {}
var Critical = func(ctx context.Context, args ...interface{}) {}

type SweetenContext interface {
	Sweeten() []zap.Field
}

func Log() *zap.Logger {
	return l
}

func Init(level string) {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()

	if err != nil {
		panic(err.Error())
	}

	l = logger

	updateLogHandlers(level)
}

func updateLogHandlers(logLevel string) {
	min := 1

	switch logLevel {
	case "debug":
		min = 0
	case "info":
		min = 1
	case "warning":
		min = 2
	case "error":
		min = 3
	case "critical":
		min = 4
	default:
		min = 1
	}

	lvls := []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel, zapcore.FatalLevel}
	lhfs := []*func(context.Context, string, ...interface{}){&Debugf, &Infof, &Warnf, &Errorf, &Criticalf}
	lhs := []*func(context.Context, ...interface{}){&Debug, &Info, &Warn, &Error, &Critical}

	for i := min; i < 5; i++ {
		*lhs[i] = makeLog(lvls[i])
		*lhfs[i] = makeLogf(lvls[i])
	}
}

func makeLog(lvl zapcore.Level) func(context.Context, ...interface{}) {
	return func(ctx context.Context, args ...interface{}) {
		logf(ctx, lvl, "", args)
	}
}

func makeLogf(lvl zapcore.Level) func(context.Context, string, ...interface{}) {
	return func(ctx context.Context, format string, args ...interface{}) {
		logf(ctx, lvl, format, args)
	}
}

func logf(ctx context.Context, lvl zapcore.Level, format string, args ...interface{}) {
	msg := ""
	if format == "" {
		msg = fmt.Sprint(args...)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	sf := ([]zap.Field)(nil)
	if ctx != nil {
		if c, ok := ctx.(SweetenContext); ok {
			sf = c.Sweeten()
		}
	}
	if ce := l.Check(lvl, msg); ce != nil {
		ce.Write(sf...)
	}
}
