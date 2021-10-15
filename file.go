package elog

import (
	"context"
	"elog/internal/filewriter"
	"io"
	"path/filepath"
)

const (
	_infoIdx = iota
	_warnIdx
	_errorIdx
	_totalIdx
)

var _fileNames = map[int]string{
	_infoIdx:  "info.log",
	_warnIdx:  "warning.log",
	_errorIdx: "error.log",
}

type FileHandle struct {
	render Render
	fws [_totalIdx]*filewriter.FileWriter
}

func NewFile(dir string, bufferSize, rorateSize int64, maxLogFile int) *FileHandle {
	newWriter := func(name string) *filewriter.FileWriter{
		var options []filewriter.Option
		w, err := filewriter.New(filepath.Join(dir, name), options...)
		if err != nil {
			panic(err)
		}
		return w
	}

	handler := &FileHandle{
		render: newPatternRender(defaultPattern),
	}

	for idx, name := range _fileNames {
		handler.fws[idx] = newWriter(name)
	}

	return handler
}

func (h *FileHandle) Log(ctx context.Context, lv Level, args ...D)  {
	d := toMap(args...)
	addExtraField(ctx, d)
	var w io.Writer
	switch lv {
	case _warnLevel:
		w = h.fws[_warnIdx]
	case _errorLevel:
		w = h.fws[_errorIdx]
	default:
		w = h.fws[_infoIdx]
	}
	h.render.Render(w, d)
	w.Write([]byte("\n"))
}

func (h *FileHandle) Close() error {
	for _, fw := range h.fws {
		fw.Close()
	}
	return nil
}