package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init() {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.JSONFormatter{})

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		parsedLevel = logrus.InfoLevel
	}
	Log.SetLevel(parsedLevel)
}

func Info(args ...any) {
	Log.Info(args...)
}

func Error(args ...any) {
	Log.Error(args...)
}

func Fatal(args ...any) {
	Log.Fatal(args...)
}

func Warn(args ...any) {
	Log.Warn(args...)
}

func Debug(args ...any) {
	Log.Debug(args...)
}
