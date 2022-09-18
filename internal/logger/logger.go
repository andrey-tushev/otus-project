package logger

import (
	"fmt"
)

type Logger struct {
	level string
}

const (
	LevelError = "error"
	LevelWarn  = "warn"
	LevelInfo  = "info"
	LevelDebug = "debug"
)

func New(level string) *Logger {
	return &Logger{
		level: level,
	}
}

func (l Logger) Error(msg string) {
	fmt.Println("ERROR: " + msg)
}

func (l Logger) Warn(msg string) {
	if l.level == LevelError || l.level == LevelWarn {
		fmt.Println("WARN: " + msg)
	}
}

func (l Logger) Info(msg string) {
	if l.level == LevelError || l.level == LevelWarn || l.level == LevelInfo {
		fmt.Println("INFO: " + msg)
	}
}

func (l Logger) Debug(msg string) {
	if l.level == LevelError || l.level == LevelWarn || l.level == LevelInfo || l.level == LevelDebug {
		fmt.Println("DEBUG: " + msg)
	}
}
