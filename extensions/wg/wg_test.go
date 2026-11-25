package wg

import (
	"strings"
	"testing"
	ical "github.com/arran4/golang-ical"
)

func TestWgProperties(t *testing.T) {
	cal := ical.NewCalendar()
	SetTimezone(cal, "Europe/Berlin")
	SetCalName(cal, "Test")
	SetCalDesc(cal, "Desc")
	SetRelCalId(cal, "ID")

	output := cal.Serialize()

	if !strings.Contains(output, "X-WG-TIMEZONE:Europe/Berlin") {
		t.Errorf("Expected output to contain X-WG-TIMEZONE, got %s", output)
	}
	if !strings.Contains(output, "X-WG-CALNAME:Test") {
		t.Errorf("Expected output to contain X-WG-CALNAME, got %s", output)
	}
	if !strings.Contains(output, "X-WG-CALDESC:Desc") {
		t.Errorf("Expected output to contain X-WG-CALDESC, got %s", output)
	}
	if !strings.Contains(output, "X-WG-RELCALID:ID") {
		t.Errorf("Expected output to contain X-WG-RELCALID, got %s", output)
	}
}
