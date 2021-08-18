package seed

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/utils"
)

type LogWriter interface {
	io.StringWriter
	Flush() error
}

type StdoutLogWriter struct {
	LogWriter
}

func (l *StdoutLogWriter) WriteString(s string) (n int, err error) {
	return os.Stdout.WriteString(s)
}

func (l *StdoutLogWriter) Flush() error {
	return os.Stdout.Sync()
}

type DefaultProgressLogger struct {
	ProgressLogger
	Out           LogWriter
	LastStep      time.Time
	LastProgress  time.Time
	Verbose       bool
	Silent        bool
	CurrentTaskID string
	ProgressStore ProgressStore
}

func NewDefaultProgressLogger(out LogWriter, silent bool, verbose bool, progress_store ProgressStore) *DefaultProgressLogger {
	if out == nil {
		out = &StdoutLogWriter{}
	}
	return &DefaultProgressLogger{Out: out, Silent: silent, Verbose: verbose, ProgressStore: progress_store, LastStep: time.Now(), LastProgress: time.Time{}}
}

func (p *DefaultProgressLogger) SetCurrentTaskID(id string) {
	p.CurrentTaskID = id
}

func (p *DefaultProgressLogger) LogMessage(msg string) {
	if p.Out != nil {
		p.Out.WriteString(fmt.Sprintf("[%s] %s\n", timestamp(), msg))
		p.Out.Flush()
	}
}

func timestamp() string {
	return time.Now().Format("15:04:05")
}

func formatBBox(bbox vec2d.Rect) string {
	return fmt.Sprintf("%.5f, %.5f, %.5f, %.5f", bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1])
}

func (p *DefaultProgressLogger) LogStep(progress *SeedProgress) {
	if !p.Verbose {
		return
	}
	if p.LastStep.Add(time.Nanosecond).Before(time.Now()) && p.Out != nil {
		p.Out.WriteString(fmt.Sprintf("[%s] %6.2f%%\t%-20s \r", timestamp(), progress.progress*100, progress.ToString()))
		p.Out.Flush()
		p.LastStep = time.Now()
	}
}

func (p *DefaultProgressLogger) LogProgress(progress *SeedProgress, level int, bbox vec2d.Rect, tiles int) {
	progressInterval := 1
	if !p.Verbose {
		progressInterval = 30
	}

	logProgess := false
	if progress.progress == 1.0 || (p.LastProgress.Add(time.Duration(progressInterval)).Before(time.Now())) {
		p.LastProgress = time.Now()
		logProgess = true
	}

	if logProgess {
		if p.ProgressStore != nil && p.CurrentTaskID != "" {
			p.ProgressStore.Store(p.CurrentTaskID, progress.CurrentProgressIdentifier())
			p.ProgressStore.Save()
		}
	}

	if p.Silent {
		return
	}

	if logProgess && p.Out != nil {
		p.Out.WriteString(fmt.Sprintf("[%s] %d %6.2f%% %s (%d tiles)\n", timestamp(), level, progress.progress*100, formatBBox(bbox), tiles))
		p.Out.Flush()
	}
}

type LocalProgressStore struct {
	ProgressStore
	filename string
	status   map[string][][2]int
}

func NewLocalProgressStore(filename string, continue_seed bool) *LocalProgressStore {
	ret := &LocalProgressStore{filename: filename}
	if continue_seed {
		ret.status = ret.Load()
	} else {
		ret.status = map[string][][2]int{}
	}
	return ret
}

func (s *LocalProgressStore) marshal() []byte {
	data, err := json.Marshal(s.status)
	if err == nil {
		return data
	}
	return nil
}

func (s *LocalProgressStore) unmarshal(data []byte) error {
	return json.Unmarshal(data, &s.status)
}

func (s *LocalProgressStore) Store(id string, progress [][2]int) {
	s.status[id] = progress
}

func (s *LocalProgressStore) Get(id string) [][2]int {
	if v, ok := s.status[id]; ok {
		return v
	}
	return nil
}

func (s *LocalProgressStore) Load() map[string][][2]int {
	if ok, err := utils.FileExists(s.filename); !ok || err != nil {
		return nil
	} else {
		f, err := os.Open(s.filename)
		defer f.Close()
		if err != nil {
			return nil
		}
		d, err := ioutil.ReadAll(f)
		if err != nil {
			return nil
		}
		err = s.unmarshal(d)
		if err != nil {
			return nil
		}
		return s.status
	}
}

func (s *LocalProgressStore) Save() error {
	data := s.marshal()
	return os.WriteFile(s.filename, data, os.ModePerm)
}

func (s *LocalProgressStore) Remove() error {
	s.status = map[string][][2]int{}
	if ok, err := utils.FileExists(s.filename); ok && err == nil {
		return os.Remove(s.filename)
	}
	return os.ErrNotExist
}
