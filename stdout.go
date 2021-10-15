package elog

import (
	"context"
	"os"
)

var _defaultStdout = NewStdout()

type StdoutHandler struct {
	render Render
}

func NewStdout() *StdoutHandler {
	return &StdoutHandler{render: newPatternRender(defaultPattern)}
}

func (h *StdoutHandler) Log(ctx context.Context, lv Level, args ...D) {
	d := toMap(args...)
	addExtraField(ctx, d)
	h.render.Render(os.Stderr, d)
	os.Stderr.Write([]byte("\n"))
}

func (h *StdoutHandler) Close() error {
	return nil
}