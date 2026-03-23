package logging

import (
	"log"
	"os"
	"strings"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

var currentLevel = parseLevel(os.Getenv("LOG_LEVEL"))

func SetLevel(level string) {
	currentLevel = parseLevel(level)
}

func Debugf(format string, args ...any) {
	logf(LevelDebug, format, args...)
}

func Infof(format string, args ...any) {
	logf(LevelInfo, format, args...)
}

func Warnf(format string, args ...any) {
	logf(LevelWarn, format, args...)
}

func Errorf(format string, args ...any) {
	logf(LevelError, format, args...)
}

func logf(level Level, format string, args ...any) {
	if !enabled(level) {
		return
	}
	log.Printf("["+string(level)+"] "+format, args...)
}

func enabled(level Level) bool {
	return levelWeight(level) >= levelWeight(currentLevel)
}

func levelWeight(level Level) int {
	switch level {
	case LevelDebug:
		return 10
	case LevelInfo:
		return 20
	case LevelWarn:
		return 30
	case LevelError:
		return 40
	default:
		return 20
	}
}

func parseLevel(level string) Level {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return LevelDebug
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}
