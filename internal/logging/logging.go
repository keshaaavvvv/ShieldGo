package logging

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Event struct {
	Time   string `json:"time"`
	IP     string `json:"ip"`
	Path   string `json:"path"`
	Action string `json:"action"`
	Status int    `json:"status"`
}

type Logger struct {
	mu sync.Mutex
	f  *os.File
}

func Open(path string) (*Logger, error) {
	f, err := os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return nil, err
	}

	return &Logger{
		f: f,
	}, nil
}

func (l *Logger) Log(
	ip string,
	path string,
	action string,
	status int,
) {
	ev := Event{
		Time:   time.Now().Format(time.RFC3339),
		IP:     ip,
		Path:   path,
		Action: action,
		Status: status,
	}

	b, err := json.Marshal(ev)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.f.Write(b)
	_, _ = l.f.Write([]byte("\n"))
}

func (l *Logger) Close() error {
	return l.f.Close()
}
