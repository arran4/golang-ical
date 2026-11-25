package xapple

import (
	ical "github.com/arran4/golang-ical"
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
