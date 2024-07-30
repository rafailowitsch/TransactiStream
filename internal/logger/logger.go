package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
)

var Log zerolog.Logger

func InitLogger() {
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	Log = zerolog.New(output).With().Timestamp().Logger()
}

func Info(msg string) {
	Log.Info().Msg(msg)
}

func Infof(msg string, args ...interface{}) {
	Log.Info().Msg(fmt.Sprintf(msg, args...))
}

func Error(msg string) {
	Log.Error().Msg(msg)
}

func Errorf(msg string, args ...interface{}) {
	Log.Error().Msg(fmt.Sprintf(msg, args...))
}
