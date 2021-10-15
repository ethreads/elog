package elog

import (
	"context"
	pkgerr "github.com/pkg/errors"
	"time"
)
const (
	_timeFormat = "2006-01-02T15:04:05.999999"

	//  log level name: INFO, WARN...
	_level = "level"
	// log time.
	_time = "time"
	// log file.
	_source = "source"
	// common log filed.
	_log = "log"
)

type Handler interface {
	// Log handler log
	// variadic D is k-v struct represent log content
	Log(context.Context, Level, ...D)

	// Close handler
	Close() error
}

func newHandlers(filters []string, handlers ...Handler) *Handlers {
	return &Handlers{handlers: handlers}
}

type Handlers struct {
	handlers []Handler
}

func (hs *Handlers) Log(ctx context.Context, lv Level, d ...D) {
	hasSource := false
	for i := range d {
		if d[i].Key == _source {
			hasSource = true
		}
	}
	if !hasSource {
		fn := funcName(3)
		d = append(d, KVString(_source, fn))
	}
	d = append(d, KV(_time, time.Now()), KVString(_level, lv.String()))
	for _, h := range hs.handlers {
		h.Log(ctx, lv, d...)
	}
}

func (hs *Handlers) Close() (err error) {
	for _, h := range hs.handlers {
		if e := h.Close(); e != nil {
			err = pkgerr.WithStack(err)
		}
	}
	return
}