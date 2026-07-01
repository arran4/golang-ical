package xapple

import (
	ical "github.com/arran4/golang-ical"
	"image/color"
	"strings"
	"testing"
)

func TestAppleProperties(t *testing.T) {
	cal := ical.NewCalendar()
	SetCalendarColor(cal, "red")
	SetRegion(cal, "US")

	event := cal.AddEvent("id")
	SetStructuredLocation(event, "geo:37.3318,-122.0312")
	SetTravelDuration(event, "PT1H")

	output := cal.Serialize()

	if !strings.Contains(output, "X-APPLE-CALENDAR-COLOR:red") {
		t.Errorf("Expected output to contain X-APPLE-CALENDAR-COLOR:red, got %s", output)
	}
	if !strings.Contains(output, "X-APPLE-REGION:US") {
		t.Errorf("Expected output to contain X-APPLE-REGION:US, got %s", output)
	}
	if !strings.Contains(output, "X-APPLE-STRUCTURED-LOCATION:geo:37.3318\\,-122.0312") {
		t.Errorf("Expected output to contain X-APPLE-STRUCTURED-LOCATION:geo:37.3318\\,-122.0312, got %s", output)
	}
	if !strings.Contains(output, "X-APPLE-TRAVEL-DURATION:PT1H") {
		t.Errorf("Expected output to contain X-APPLE-TRAVEL-DURATION:PT1H, got %s", output)
	}
}

func TestCalendarColorFromColor(t *testing.T) {
	cal := ical.NewCalendar()
	SetCalendarColorFromColor(cal, color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 255})

	if got := GetCalendarColorAsString(cal); got != "#123456" {
		t.Fatalf("got %s, want #123456", got)
	}

	c, err := GetCalendarColorAsColor(cal)
	if err != nil {
		t.Fatalf("GetCalendarColorAsColor() error = %v", err)
	}
	if c != (color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 255}) {
		t.Fatalf("got %v, want #123456 as color.RGBA", c)
	}
}

func TestParsedCalendarColor(t *testing.T) {
	cal, err := ical.ParseCalendar(strings.NewReader("BEGIN:VCALENDAR\nVERSION:2.0\nX-APPLE-CALENDAR-COLOR:#123456\nEND:VCALENDAR"))
	if err != nil {
		t.Fatalf("ParseCalendar() error = %v", err)
	}

	p := GetCalendarColor(cal)
	if p == nil {
		t.Fatal("expected X-APPLE-CALENDAR-COLOR property")
	}

	c, err := GetCalendarColorAsColor(cal)
	if err != nil {
		t.Fatalf("GetCalendarColorAsColor() error = %v", err)
	}
	if c != (color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 255}) {
		t.Fatalf("got %v, want #123456 as color.RGBA", c)
	}
}

func TestCalendarColorMissing(t *testing.T) {
	if _, err := GetCalendarColorAsColor(ical.NewCalendar()); err == nil {
		t.Fatal("expected missing calendar color error")
	}
}
