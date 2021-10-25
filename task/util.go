package task

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
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
	CurrentTaskId string
	ProgressStore ProgressStore
}

func NewDefaultProgressLogger(out LogWriter, silent bool, verbose bool, progress_store ProgressStore) *DefaultProgressLogger {
	if out == nil {
		out = &StdoutLogWriter{}
	}
	return &DefaultProgressLogger{Out: out, Silent: silent, Verbose: verbose, ProgressStore: progress_store, LastStep: time.Now(), LastProgress: time.Time{}}
}

func (p *DefaultProgressLogger) GetStore() ProgressStore {
	return p.ProgressStore
}

func (p *DefaultProgressLogger) SetCurrentTaskId(id string) {
	p.CurrentTaskId = id
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

func (p *DefaultProgressLogger) LogStep(progress *TaskProgress) {
	if !p.Verbose {
		return
	}
	if p.LastStep.Add(time.Nanosecond).Before(time.Now()) && p.Out != nil {
		p.Out.WriteString(fmt.Sprintf("[%s] %6.2f%%\t%-20s \r", timestamp(), progress.progress*100, progress.ToString()))
		p.Out.Flush()
		p.LastStep = time.Now()
	}
}

func (p *DefaultProgressLogger) LogProgress(progress *TaskProgress, level int, bbox vec2d.Rect, tiles int) {
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
		if p.ProgressStore != nil && p.CurrentTaskId != "" {
			p.ProgressStore.Store(p.CurrentTaskId, progress.CurrentProgressIdentifier())
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
	status   map[string]interface{}
}

func NewLocalProgressStore(filename string, continue_seed bool) *LocalProgressStore {
	ret := &LocalProgressStore{filename: filename}
	if continue_seed {
		ret.status = ret.load()
	} else {
		ret.status = map[string]interface{}{}
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

func (s *LocalProgressStore) Store(id string, progress interface{}) {
	s.status[id] = progress
	s.flush()
}

func (s *LocalProgressStore) Get(id string) interface{} {
	if v, ok := s.status[id]; ok {
		return v
	}
	return nil
}

func (s *LocalProgressStore) load() map[string]interface{} {
	if !utils.FileExists(s.filename) {
		return nil
	} else {
		f, err := os.Open(s.filename)
		if err != nil {
			return nil
		}
		defer f.Close()
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

func (s *LocalProgressStore) flush() error {
	data := s.marshal()
	return os.WriteFile(s.filename, data, os.ModePerm)
}

func (s *LocalProgressStore) remove() error {
	s.status = map[string]interface{}{}
	if utils.FileExists(s.filename) {
		return os.Remove(s.filename)
	}
	return os.ErrNotExist
}

func izip_longest(fillvalue interface{}, iterables ...interface{}) [][]interface{} {
	if len(iterables) == 0 {
		return nil
	}

	s := reflect.ValueOf(iterables[0])
	size := s.Len()
	for _, v := range iterables[1:] {
		s_ := reflect.ValueOf(v)
		if s_.Len() > size {
			size = s_.Len()
		}
	}

	results := [][]interface{}{}

	for i := 0; i < size; i += 1 {
		newresult := make([]interface{}, len(iterables))
		for j, v := range iterables {
			s_ := reflect.ValueOf(v)
			if i < s_.Len() {
				newresult[j] = s_.Index(i).Interface()
			} else {
				newresult[j] = fillvalue
			}
		}

		results = append(results, newresult)
	}

	return results
}

func reverse(slice interface{}) interface{} {
	s := reflect.ValueOf(slice)

	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	swp := reflect.Swapper(s.Interface())
	for i, j := 0, s.Len()-1; i < j; i, j = i+1, j-1 {
		swp(i, j)
	}
	return slice
}
