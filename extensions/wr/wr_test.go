package wr

import (
	"strings"
	"testing"
	ical "github.com/arran4/golang-ical"
)

func TestWrProperties(t *testing.T) {
	cal := ical.NewCalendar()
	SetTimezone(cal, "Europe/Berlin")
	SetCalName(cal, "Test")
	SetCalDesc(cal, "Desc")
	SetRelCalId(cal, "ID")

	output := cal.Serialize()

	if !strings.Contains(output, "X-WR-TIMEZONE:Europe/Berlin") {
		t.Errorf("Expected output to contain X-WR-TIMEZONE, got %s", output)
	}
	if !strings.Contains(output, "X-WR-CALNAME:Test") {
		t.Errorf("Expected output to contain X-WR-CALNAME, got %s", output)
	}
	if !strings.Contains(output, "X-WR-CALDESC:Desc") {
		t.Errorf("Expected output to contain X-WR-CALDESC, got %s", output)
	}
	if !strings.Contains(output, "X-WR-RELCALID:ID") {
		t.Errorf("Expected output to contain X-WR-RELCALID, got %s", output)
	}
}
