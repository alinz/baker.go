package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Level int

const (
	Debug Level = iota
	Warn
	Info
	Error
)

type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Error(format string, args ...interface{})
	Level(level Level)
}

type Simple struct {
	level    Level
	internal *log.Logger
}

var _ (Logger) = (*Simple)(nil)

func (s *Simple) log(prefix string, format string, args []interface{}) {
	format = fmt.Sprintf("%s: %s", prefix, format)
	s.internal.Printf(format, args...)
}

func (s *Simple) Info(format string, args ...interface{}) {
	if Info >= s.level {
		s.log("INFO", format, args)
	}
}

func (s *Simple) Warn(format string, args ...interface{}) {
	if Warn >= s.level {
		s.log("WARN", format, args)
	}
}

func (s *Simple) Debug(format string, args ...interface{}) {
	if Debug >= s.level {
		s.log("DEBUG", format, args)
	}
}

func (s *Simple) Error(format string, args ...interface{}) {
	if Error >= s.level {
		s.log("ERROR", format, args)
	}
}

func (s *Simple) Level(level Level) {
	s.level = level
}

func NewSimple(out io.Writer) *Simple {
	if out == nil {
		out = os.Stdout
	}

	return &Simple{
		level:    Debug,
		internal: log.New(out, "", 0),
	}
}

var Default = NewSimple(nil)
