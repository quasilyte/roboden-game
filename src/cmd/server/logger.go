package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type logger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	GetSize() int64
	Truncate() error
}

type fileLogger struct {
	mu sync.RWMutex
	f  *os.File

	filename string
	size     int64
}

func (l *fileLogger) GetSize() int64 {
	return atomic.LoadInt64(&l.size)
}

func (l *fileLogger) Truncate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.filename == "" {
		return nil
	}
	if err := l.f.Close(); err != nil {
		return err
	}
	f, err := os.Create(l.filename)
	if err != nil {
		return err
	}
	l.f = f
	return nil
}

func (l *fileLogger) timeNow() string {
	now := time.Now()
	return fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
}

func (l *fileLogger) Info(format string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var n int
	if len(args) == 0 {
		n, _ = fmt.Fprintln(l.f, "[info] "+l.timeNow()+" "+format)
	} else {
		n, _ = fmt.Fprintf(l.f, "[info] "+l.timeNow()+" "+format+"\n", args...)
	}
	atomic.AddInt64(&l.size, int64(n))
}

func (l *fileLogger) Error(format string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var n int
	if len(args) == 0 {
		n, _ = fmt.Fprintln(l.f, "[error] "+l.timeNow()+" "+format)
	} else {
		n, _ = fmt.Fprintf(l.f, "[error] "+l.timeNow()+" "+format+"\n", args...)
	}
	atomic.AddInt64(&l.size, int64(n))
}
