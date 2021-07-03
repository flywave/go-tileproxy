package debug

import (
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"
)

type LogDebugger struct {
	Output  io.Writer
	Prefix  string
	Flag    int
	logger  *log.Logger
	counter int32
	start   time.Time
}

func (l *LogDebugger) Init() error {
	l.counter = 0
	l.start = time.Now()
	if l.Output == nil {
		l.Output = os.Stderr
	}
	l.logger = log.New(l.Output, l.Prefix, l.Flag)
	return nil
}

func (l *LogDebugger) Event(e *Event) {
	i := atomic.AddInt32(&l.counter, 1)
	l.logger.Printf("[%06d] %d [%6d - %s] %q (%s)\n", i, e.CollectorID, e.RequestID, e.Type, e.Values, time.Since(l.start))
}
