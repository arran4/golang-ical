package ics

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestResolveTimezone_IANA(t *testing.T) {
	loc, err := resolveTimezone("America/New_York")
	if err != nil {
		t.Fatalf("expected no error for IANA tz, got %v", err)
	}
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
}

func TestResolveTimezone_Windows(t *testing.T) {
	loc, err := resolveTimezone("New Zealand Standard Time", WindowsTimezoneToIANA)
	if err != nil {
		t.Fatalf("expected no error for Windows tz, got %v", err)
	}
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
	auckland, _ := time.LoadLocation("Pacific/Auckland")
	now := time.Now()
	if now.In(loc).String() != now.In(auckland).String() {
		t.Errorf("expected Pacific/Auckland equivalent, got %v", loc)
	}
}

func TestResolveTimezone_WindowsDisabled(t *testing.T) {
	_, err := resolveTimezone("New Zealand Standard Time")
	if err == nil {
		t.Fatal("expected error for Windows tz without mapping enabled")
	}
}

func TestResolveTimezone_WindowsFalse(t *testing.T) {
	_, err := resolveTimezone("New Zealand Standard Time", func(string) *time.Location { return nil })
	if err == nil {
		t.Fatal("expected error for Windows tz with mapping explicitly disabled")
	}
}

func TestResolveTimezone_Unknown(t *testing.T) {
	_, err := resolveTimezone("Fake/Timezone")
	if err == nil {
		t.Fatal("expected error for unknown tz")
	}
}

func TestResolveTimezone_NilOption(t *testing.T) {
	_, err := resolveTimezone("America/New_York", nil)
	if err == nil {
		t.Fatal("expected error for nil variadic option")
	}
	if !errors.Is(err, ErrInvalidOpArg) {
		t.Fatalf("expected ErrInvalidOpArg, got %v", err)
	}
}

func TestResolveTimezone_MapperChain(t *testing.T) {
	loc, err := resolveTimezone(
		"New Zealand Standard Time",
		func(string) *time.Location {
			return WindowsTimezoneToIANA("New Zealand Standard Time")
		},
		func(tzid string) *time.Location {
			if tzid == "Pacific/Auckland" {
				utc, _ := time.LoadLocation("UTC")
				return utc
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("expected no error for chained mappers, got %v", err)
	}
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
	if loc.String() != "UTC" {
		t.Fatalf("expected final mapped location to win, got %v", loc)
	}
}

func TestParseTimeValue_WindowsTZID(t *testing.T) {
	icsData := "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nDTSTART;TZID=New Zealand Standard Time:20260407T120000\r\nDTEND;TZID=New Zealand Standard Time:20260407T130000\r\nSUMMARY:Test Event\r\nEND:VEVENT\r\nEND:VCALENDAR"
	cal, err := ParseCalendarWithOptions(strings.NewReader(icsData), WithWindowsTimezoneMapping())
	if err != nil {
		t.Fatalf("ParseCalendar error: %v", err)
	}
	events := cal.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	start, err := events[0].GetStartAt()
	if err != nil {
		t.Fatalf("GetStartAt error: %v", err)
	}
	if start.IsZero() {
		t.Fatal("GetStartAt returned zero time")
	}
	if start.Month() != 4 || start.Day() != 7 || start.Hour() != 12 {
		t.Errorf("expected 2026-04-07T12:00, got %v", start)
	}
}

func TestWithWindowsTimezoneMapping(t *testing.T) {
	icsData := "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nDTSTART;TZID=China Standard Time:20260407T090000\r\nSUMMARY:Test Event\r\nEND:VEVENT\r\nEND:VCALENDAR"

	cal, err := ParseCalendarWithOptions(strings.NewReader(icsData), WithWindowsTimezoneMapping())
	if err != nil {
		t.Fatalf("ParseCalendarWithOptions error: %v", err)
	}
	events := cal.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	start, err := events[0].GetStartAt()
	if err != nil {
		t.Fatalf("GetStartAt error: %v", err)
	}
	if start.IsZero() {
		t.Fatal("GetStartAt returned zero time")
	}
	if start.Month() != 4 || start.Day() != 7 || start.Hour() != 9 {
		t.Errorf("expected 2026-04-07T09:00, got %v", start)
	}
}

func TestSerializeTimezoneMapping(t *testing.T) {
	cal := NewCalendar()
	cal.AddTimezone("Pacific/Auckland")
	event := cal.AddEvent("event-1")
	event.SetProperty(ComponentPropertyDtStart, "20260407T090000", WithTZID("Pacific/Auckland"))

	serialized := cal.Serialize(WithWindowsTimezoneMappingForSerialization())
	if !strings.Contains(serialized, "TZID:New Zealand Standard Time") {
		t.Fatalf("expected timezone component to serialize as Windows timezone, got %q", serialized)
	}
	if !strings.Contains(serialized, "DTSTART;TZID=New Zealand Standard Time:20260407T090000") {
		t.Fatalf("expected DTSTART TZID parameter to serialize as Windows timezone, got %q", serialized)
	}
}
