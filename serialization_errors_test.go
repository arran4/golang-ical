package ics

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

type failingSerializationWriter struct {
	bytes.Buffer
	remaining int
	err       error
	failed    bool
	after     int
}

func (w *failingSerializationWriter) Write(p []byte) (int, error) {
	if w.failed {
		w.after++
		return w.Buffer.Write(p)
	}
	if len(p) > w.remaining {
		n, _ := w.Buffer.Write(p[:w.remaining])
		w.failed = true
		return n, w.err
	}
	w.remaining -= len(p)
	return w.Buffer.Write(p)
}

func (w *failingSerializationWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func TestCalendarSerializeToWriteErrors(t *testing.T) {
	cal := NewCalendar()
	event := cal.AddEvent("event-id")
	event.SetSummary("Meeting")
	alarm := NewAlarm("")
	alarm.SetAction(ActionDisplay)
	alarm.SetDescription("Reminder")
	event.AddVAlarm(alarm)
	want := cal.Serialize(WithNewLineWindows)
	writeErr := errors.New("output unavailable")

	// Fail at every byte, including component boundaries and the final line.
	for limit := 0; limit < len(want); limit++ {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			w := &failingSerializationWriter{remaining: limit, err: writeErr}
			if err := cal.SerializeTo(w, WithNewLineWindows); !errors.Is(err, writeErr) {
				t.Errorf("SerializeTo error = %v, want %v", err, writeErr)
			}
			if w.after != 0 {
				t.Errorf("serialization made %d writes after the first error", w.after)
			}
			if got := w.String(); got != want[:limit] {
				t.Errorf("output = %q, want prefix %q", got, want[:limit])
			}
		})
	}
}
