package tracing

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

var Enabled bool = false
var Target io.Writer = os.Stdout

var Detailed bool = false

func Run(core func(), pre, post func(w io.Writer)) {
	wasEnabled := Enabled
	if wasEnabled {
		pre(Target)
	}
	core()
	if wasEnabled {
		post(Target)
	}
}

const (
	BeginEvent = "B"
	EndEvent   = "E"
)

var firstEvent bool = true

func WriteJSON(w io.Writer, event string, t time.Time, pid, tid int, name string, args map[string]interface{}) error {
	if firstEvent {
		w.Write([]byte(" "))
	} else {
		w.Write([]byte(","))
	}
	firstEvent = false
	value := struct {
		Name            string                 `json:"name"`
		Phase           string                 `json:"ph"`
		TimestampMicros int64                  `json:"ts"`
		Pid             int                    `json:"pid"`
		Tid             int                    `json:"tid"`
		Args            map[string]interface{} `json:"args,omitempty"`
	}{
		Name:            name,
		Phase:           event,
		TimestampMicros: t.UnixNano() / 1000,
		Pid:             pid,
		Tid:             tid,
		Args:            args,
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}
