package logger

import (
	"bufio"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type FlushTimerBuff struct {
	*bufio.Writer
	m            sync.RWMutex
	w            io.Writer
	cfg          OutputConfig
	stopSignal   chan bool
	notifySignal chan error
}

func NewFileWriter(cfg OutputConfig) (*FlushTimerBuff, error) {
	obj := &FlushTimerBuff{
		cfg:          cfg,
		stopSignal:   make(chan bool),
		notifySignal: make(chan error),
	}

	if len(cfg.Path) > 0 {
		fileWriter, writer, err := openFile(cfg.Path, cfg.BufferSize)
		if err != nil {
			return nil, err
		}
		obj.w = fileWriter
		obj.Writer = writer
	}

	return obj, nil
}

func openFile(path string, bufferSize int) (*os.File, *bufio.Writer, error) {
	fileWriter, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}
	writer := bufio.NewWriterSize(fileWriter, bufferSize)
	return fileWriter, writer, nil
}

func (f *FlushTimerBuff) Write(p []byte) (nn int, err error) {
	if f.Writer == nil {
		return
	}
	f.m.Lock()
	nn, err = f.Writer.Write(p)
	f.m.Unlock()
	return nn, err
}

func (f *FlushTimerBuff) Flush() (err error) {
	f.m.Lock()
	defer f.m.Unlock()
	return f.flush()
}

func (f *FlushTimerBuff) flush() (err error) {
	if f.Writer == nil {
		return
	}
	if f.Writer.Buffered() == 0 {
		return
	}
	err = f.Writer.Flush()
	f.Writer.Reset(f.w)
	return
}

func (f *FlushTimerBuff) Sync() error {
	return f.FileFlashWorkerRestart()
}

func (f *FlushTimerBuff) FileFlashWorkerRestart() (err error) {
	if f.Writer == nil {
		return
	}

	f.m.Lock()
	defer f.m.Unlock()

	err = f.flush()
	if err != nil {
		return err
	}
	err = f.close()
	if err != nil {
		return err
	}
	fileWriter, writer, err := openFile(f.cfg.Path, f.cfg.BufferSize)
	if err != nil {
		return err
	}
	f.w = fileWriter
	f.Writer = writer
	return nil
}

func (f *FlushTimerBuff) close() error {
	return f.w.(*os.File).Close()
}

func (f *FlushTimerBuff) FileFlashWorker() {
	if f.Writer == nil {
		return
	}

	var seconds int64 = 5
	if f.cfg.FlushSeconds > 0 {
		seconds = f.cfg.FlushSeconds
	}
	var t = time.NewTicker(time.Second * time.Duration(seconds))
	go func() {
		defer t.Stop()
		for {
			select {
			case <-f.stopSignal:
				f.notifySignal <- f.Flush()
				err := f.close()
				if err != nil {
					log.Println("log close file error: " + err.Error())
				}
				return
			case <-t.C:
				if err := f.Flush(); err != nil {
					log.Println("log flush error: " + err.Error())
				}
			}
		}
	}()
}

func (f *FlushTimerBuff) FileFlashWorkerStop() (err error) {
	if f.Writer == nil {
		return
	}

	go func() {
		f.stopSignal <- true
	}()

	err = <-f.notifySignal

	return err
}
