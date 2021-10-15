package elog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func initStdout() {
	conf := &Config{
		Stdout: true,
	}
	Init(conf)
}

func testLog(t *testing.T) {
	t.Run("Info", func(t *testing.T) {
		Info("hello %s", "world")
	})
}

func TestStout(t *testing.T) {
	initStdout()
	testLog(t)
	assert.Equal(t, nil, Close())
}

func initFile() {
	conf := &Config{
		Dir: "/Users/easyboom/go/src/elog/log",
	}
	Init(conf)
}

func TestFile(t *testing.T) {
	initFile()
	testLog(t)
	assert.Equal(t, nil, Close())
}