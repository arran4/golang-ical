package ics

import (
	"strings"
	"testing"
)

func TestCalendarAttachment(t *testing.T) {
	cal := NewCalendar()
	event := cal.AddEvent("test-event")
	event.AddAttachment("http://example.com/attachment.txt", WithFmtType("text/plain"))

	serialized := cal.Serialize()
	if !strings.Contains(serialized, "ATTACH;FMTTYPE=text/plain;VALUE=URI:http://example.com/attachment.txt") {
		t.Errorf("Serialized calendar does not contain the expected ATTACH property with VALUE=URI")
	}
}
