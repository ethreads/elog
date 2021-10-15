package filewriter

import "time"

const (
	RotateDaily = "2006-01-02"
)

var defaultOption = option{
	RotateFormat:   RotateDaily,
	MaxSize:        1 << 30,
	ChanSize:       1024 * 8,
	RotateInterval: 10 * time.Second,
}

type option struct {
	RotateFormat string
	MaxFile      int
	MaxSize      int64
	ChanSize     int

	WriteTimeout time.Duration
	RotateInterval time.Duration
}

type Option func(opt *option)
