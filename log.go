package elog

import (
	"context"
	"fmt"
	"io"
)

type Config struct {
	// stdout
	Stdout bool

	// file
	Dir string
	// buffer size
	FileBufferSize int64
	// MaxLogFile
	MaxLogFile int
	// RotateSize
	RotateSize int64

	// Module=""
	// The syntax of the argument is a map of pattern=N,
	// where pattern is a literal file name (minus the ".go" suffix) or
	// "glob" pattern and N is a V level. For instance:
	// [module]
	//   "service" = 1
	//   "dao*" = 2
	// sets the V level to 2 in all Go files whose names begin "dao".
	Module map[string]int32
}

type Render interface {
	Render(io.Writer, map[string]interface{}) error
	RenderString(map[string]interface{}) string
}

var (
	h Handler
	c *Config
)

// Init create logger with context.
func Init(conf *Config) {
	var hs []Handler
	if conf.Stdout {
		hs = append(hs, NewStdout())
	}
	if conf.Dir != "" {
		hs = append(hs, NewFile(conf.Dir, conf.FileBufferSize, conf.RotateSize, conf.MaxLogFile))
	}
	h = newHandlers([]string{}, hs...)
	c = conf
}

// Info logs a message at the info log level.
func Info(format string, args ...interface{}) {
	h.Log(context.Background(), _infoLevel, KVString(_log, fmt.Sprintf(format, args...)))
}

// Warn logs a message at the warn log level.
func Warn(format string, args ...interface{}) {}

// Error logs a message at the error log level.
func Error(format string, args ...interface{}) {
	h.Log(context.Background(), _errorLevel, KVString(_log, fmt.Sprintf(format, args...)))
}

// Close close resource
func Close() (err error) {
	err = h.Close()
	h = _defaultStdout
	return
}