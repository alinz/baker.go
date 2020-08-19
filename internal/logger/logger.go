package logger

import "log"

type Level int

const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel
	AllLevel
)

type Logger struct {
	level Level
}

func (l Logger) Info(format string, args ...interface{}) {
	if l.level >= InfoLevel {
		log.Printf("INFO: "+format, args...)
	}
}

func (l Logger) Debug(format string, args ...interface{}) {
	if l.level >= DebugLevel {
		log.Printf("DEBUG: "+format, args...)
	}
}

func (l Logger) Warn(format string, args ...interface{}) {
	if l.level >= WarnLevel {
		log.Printf("WARN: "+format, args...)
	}
}

func (l Logger) Error(format string, args ...interface{}) {
	if l.level >= ErrorLevel {
		log.Printf("ERROR: "+format, args...)
	}
}

func (l *Logger) Level(level Level) {
	l.level = level
}

var Default = Logger{
	level: AllLevel,
}

func Info(format string, args ...interface{}) {
	Default.Info(format, args...)
}

func Debug(format string, args ...interface{}) {
	Default.Debug(format, args...)
}

func Warn(format string, args ...interface{}) {
	Default.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	Default.Error(format, args...)
}
