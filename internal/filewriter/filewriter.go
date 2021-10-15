package filewriter

import (
	"bytes"
	"container/list"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type FileWriter struct {
	opt option
	dir string
	fname string

	ch chan *bytes.Buffer
	stdlog *log.Logger
	pool *sync.Pool

	lastRotateFormat string
	lastSplitNum     int

	current *wrapFile
	files   *list.List

	closed int32
	wg sync.WaitGroup
}

type rotateItem struct {
	rotateTime int64
	rotateNum int
	fname string
}

type wrapFile struct {
	fsize int64
	fp *os.File
}

func (w *wrapFile) write(p []byte) (n int, err error) {
	n, err = w.fp.Write(p)
	w.fsize += int64(n)
	return
}

func newWrapFile(fpath string) (*wrapFile, error) {
	fp, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	return &wrapFile{fp: fp, fsize: fi.Size()}, nil
}

func New(fpath string, fns ... Option) (*FileWriter, error) {
	opt := defaultOption

	fname := filepath.Base(fpath)
	if fname == "" {
		return nil, fmt.Errorf("filename can't empty")
	}
	dir := filepath.Dir(fpath)
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		return nil, fmt.Errorf("%s already exists and not a directory", dir)
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create dir %s error: %s", dir, err)
		}
	}

	current, err := newWrapFile(fpath)
	if err != nil {
		return nil, err
	}

	stdlog := log.New(os.Stderr, "flog ", log.LstdFlags)
	ch := make(chan *bytes.Buffer, opt.ChanSize)

	fw := &FileWriter{
		opt: opt,
		dir: fpath,
		ch: ch,
		stdlog: stdlog,
		pool: &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
		current: current,
	}

	fw.wg.Add(1)
	go fw.daemon()

	return fw, nil
}

func (f *FileWriter) Write(p []byte) (int, error) {
	if atomic.LoadInt32(&f.closed) == 1 {
		f.stdlog.Printf("%s", p)
		return 0, fmt.Errorf("filewriter already closed")
	}

	buf := f.getBuf()
	buf.Write(p)

	if f.opt.WriteTimeout == 0 {
		select {
		case f.ch <- buf:
			return len(p), nil
		default:
			return 0, fmt.Errorf("log channel is full")
		}
	}

	timeout := time.NewTicker(f.opt.WriteTimeout)
	select {
	case f.ch <- buf:
		return len(p), nil
	case <- timeout.C:
		return 0, fmt.Errorf("log channel is full")
	}
}

func (f *FileWriter) getBuf() *bytes.Buffer {
	return f.pool.Get().(*bytes.Buffer)
}

func (f *FileWriter) putBuf(buf *bytes.Buffer) {
	buf.Reset()
	f.pool.Put(buf)
}

func (f *FileWriter) daemon() {
	aggsbuf := &bytes.Buffer{}
	tick := time.NewTicker(f.opt.RotateInterval)
	aggstk := time.NewTicker(10 * time.Millisecond)
	var err error
	for {
		select {
		case t := <- tick.C:
			f.checkRotate(t)
		case buf, ok := <-f.ch:
			if ok {
				aggsbuf.Write(buf.Bytes())
				f.putBuf(buf)
			}
		case <-aggstk.C:
			if aggsbuf.Len() > 0 {
				if err = f.write(aggsbuf.Bytes()); err != nil {
					f.stdlog.Printf("write log error:%s", err)
				}
				aggsbuf.Reset()
			}
		}
		if atomic.LoadInt32(&f.closed) != 1 {
			continue
		}

		if err = f.write(aggsbuf.Bytes()); err != nil {
			f.stdlog.Printf("write log error: %s", err)
		}

		for buf := range f.ch {
			if err = f.write(buf.Bytes()); err != nil {
				f.stdlog.Printf("write log error: %s", err)
			}
			f.putBuf(buf)
		}
		break
	}
	f.wg.Done()
}

func (f *FileWriter) write(p []byte) error {
	if f.current == nil {
		f.stdlog.Printf("can't write log to file")
		f.stdlog.Printf("%s", p)
	}

	_, err := f.current.write(p)
	return err
}

func (f *FileWriter) checkRotate(t time.Time) {
	formatFname := func(format string, num int) string {
		if num == 0 {
			return fmt.Sprintf("%s.%s", f.fname, format)
		}
		return fmt.Sprintf("%s.%d", f.fname, num)
	}

	format := t.Format(f.opt.RotateFormat)

	// 最多保留文件
	if f.opt.MaxFile != 0 {
		for f.files.Len() > f.opt.MaxFile {
			rt := f.files.Remove(f.files.Front()).(rotateItem)
			fpath := path.Join(f.dir, rt.fname)
			if err := os.Remove(fpath); err != nil {
				f.stdlog.Printf("remove file %s error: %s", fpath, err)
			}
		}
	}

	if format != f.lastRotateFormat || (f.opt.MaxSize != 0 && f.current.fsize > f.opt.MaxSize) {
		var err error
		if err = f.current.fp.Close(); err != nil {
			f.stdlog.Printf("close current file error:%s", err)
		}

		fname := formatFname(f.lastRotateFormat, f.lastSplitNum)
		oldpath := filepath.Join(f.dir, f.fname)
		newpath := filepath.Join(f.dir, fname)
		if err = os.Rename(oldpath, newpath); err != nil {
			f.stdlog.Printf("rename file %s to %s error: %s", oldpath, newpath, err)
			return
		}

		f.current, err = newWrapFile(filepath.Join(f.dir, f.fname))
		if err != nil {
			f.stdlog.Printf("create log file error: %s", err)
		}
	}
}

func (f *FileWriter) Close() error {
	atomic.StoreInt32(&f.closed, 1)
	close(f.ch)
	f.wg.Wait()
	return nil
}