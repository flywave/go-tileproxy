package debug

type Event struct {
	Type        string
	RequestID   uint32
	CollectorID uint32
	Values      map[string]string
}

type Debugger interface {
	Init() error
	Event(e *Event)
}
