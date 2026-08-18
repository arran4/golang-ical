package ics

import (
	"errors"
	"strings"
	"testing"
)

// TestParseCalendarDeeplyNestedComponents ensures that a calendar consisting of
// many nested BEGIN components is rejected with an error instead of crashing the
// program with a fatal stack overflow via unbounded recursion.
func TestParseCalendarDeeplyNestedComponents(t *testing.T) {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	for i := 0; i < maxComponentNestingDepth+5000; i++ {
		b.WriteString("BEGIN:VEVENT\r\n")
	}

	_, err := ParseCalendar(strings.NewReader(b.String()))
	if err == nil {
		t.Fatal("expected an error for deeply nested components, got nil")
	}
	if !errors.Is(err, ErrComponentNestingTooDeep) {
		t.Fatalf("expected ErrComponentNestingTooDeep, got %v", err)
	}
}

// TestParseCalendarNormalNestingAllowed ensures the depth guard does not reject
// legitimately nested calendars (VCALENDAR > VEVENT > VALARM).
func TestParseCalendarNormalNestingAllowed(t *testing.T) {
	const input = "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:x@example.com\r\n" +
		"BEGIN:VALARM\r\n" +
		"ACTION:DISPLAY\r\n" +
		"END:VALARM\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"
	cal, err := ParseCalendar(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error parsing normally nested calendar: %v", err)
	}
	if len(cal.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(cal.Components))
	}
}
