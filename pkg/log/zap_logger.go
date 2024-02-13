package log

import (
	"fmt"
	"log"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WrapLog struct {
	l *zap.Logger
}

func (w *WrapLog) GetInstance() *zap.Logger {
	return w.l
}

func NewWrapLogger(logLevel Level, isLocalEnv bool) *WrapLog {
	return &WrapLog{
		l: NewZapLogger(logLevel.String(), isLocalEnv),
	}
}

func NewZapLogger(logLevel string, isLocalEnv bool) *zap.Logger {
	var (
		zapLogger *zap.Logger
		zapLogLvl zapcore.Level
	)

	err := zapLogLvl.Set(logLevel)
	if err != nil {
		log.Println("cannot parse logLevel, err:", err.Error())
		zapLogLvl = zap.WarnLevel
	}

	if isLocalEnv {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Level = zap.NewAtomicLevelAt(zapLogLvl)
		config.DisableStacktrace = true
		zapLogger, err = config.Build()
		if err != nil {
			log.Println("cannot build logger, err:", err.Error())
		}
		return zapLogger
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapLogLvl && lvl < zapcore.ErrorLevel
	})
	consoleInfos := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	// Configure console output.
	consoleEncoder := newJSONEncoder()
	// Join the outputs, encoders, and level-handling functions into
	// zapcore.
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleInfos, lowPriority),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(&lumberjack.Logger{
			Filename:   "./log.json",
			MaxSize:    50, // megabytes
			MaxBackups: 30,
			MaxAge:     28, // days
		}), zapcore.DebugLevel),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	zapLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.DPanicLevel))
	zap.RedirectStdLog(zapLogger)

	return zapLogger
}

// Create a new JSON log encoder with the correct settings.
func newJSONEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    appendLogLevel,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func appendLogLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("debug")
	case zapcore.InfoLevel:
		enc.AppendString("info")
	case zapcore.WarnLevel:
		enc.AppendString("warning")
	case zapcore.ErrorLevel:
		enc.AppendString("error")
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		enc.AppendString("critical")
	default:
		enc.AppendString(fmt.Sprintf("Level(%d)", l))
	}
}
