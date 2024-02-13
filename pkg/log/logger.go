package log

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	InfoLevel   Level = zap.InfoLevel   // 0, default level
	WarnLevel   Level = zap.WarnLevel   // 1
	ErrorLevel  Level = zap.ErrorLevel  // 2
	DPanicLevel Level = zap.DPanicLevel // 3, used in development log
	// PanicLevel logs a message, then panics
	PanicLevel Level = zap.PanicLevel // 4
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = zap.FatalLevel // 5
	DebugLevel Level = zap.DebugLevel // -1
)

type Field = zap.Field

var (
	Skip       = zap.Skip
	Binary     = zap.Binary
	Bool       = zap.Bool
	ByteString = zap.ByteString
	Complex128 = zap.Complex128

	Complex64 = zap.Complex64

	Float64 = zap.Float64
	Float32 = zap.Float32
	Int     = zap.Int
	Int64   = zap.Int64

	Int32     = zap.Int32
	Int16     = zap.Int16
	Int8      = zap.Int8
	String    = zap.String
	Uint      = zap.Uint
	Uint64    = zap.Uint64
	Uint32    = zap.Uint32
	Uint16    = zap.Uint16
	Uint8     = zap.Uint8
	Uintptr   = zap.Uintptr
	Reflect   = zap.Reflect
	Namespace = zap.Namespace
	Stringer  = zap.Stringer
	Time      = zap.Time
	Stack     = zap.Stack
	Duration  = zap.Duration
	Any       = zap.Any
)

func init() {
	w = NewWrapLogger(DebugLevel, false)
}

func GetLogLevel(l string) Level {
	var zapLogLvl zapcore.Level
	err := zapLogLvl.Set(l)
	if err != nil {
		log.Println("cannot parse logLevel, err:", err.Error())
		zapLogLvl = zap.ErrorLevel
	}
	return zapLogLvl
}

func Setup(environment string, l string) {
	w = NewWrapLogger(GetLogLevel(l), environment == "local")
}

func With(fields ...Field) {
	w.l.With(fields...)
}

func Debug(msg string, fields ...Field) {
	w.l.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	w.l.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	w.l.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	w.l.Error(msg, fields...)
}
func DPanic(msg string, fields ...Field) {
	w.l.DPanic(msg, fields...)
}
func Panic(msg string, fields ...Field) {
	w.l.Panic(msg, fields...)
}
func Fatal(msg string, fields ...Field) {
	w.l.Fatal(msg, fields...)
}
