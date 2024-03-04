package ics

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetDuration(t *testing.T) {
	date, _ := time.Parse(time.RFC822, time.RFC822)
	duration := time.Duration(float64(time.Hour) * 2)

	testCases := []struct {
		name   string
		start  time.Time
		end    time.Time
		output string
	}{
		{
			name:  "test set duration - start",
			start: date,
			output: `BEGIN:VEVENT
UID:test-duration
DTSTART:20060102T150400Z
DTEND:20060102T170400Z
END:VEVENT
`,
		},
		{
			name: "test set duration - end",
			end:  date,
			output: `BEGIN:VEVENT
UID:test-duration
DTEND:20060102T150400Z
DTSTART:20060102T130400Z
END:VEVENT
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEvent("test-duration")
			if !tc.start.IsZero() {
				e.SetStartAt(tc.start)
			}
			if !tc.end.IsZero() {
				e.SetEndAt(tc.end)
			}
			err := e.SetDuration(duration)

			// we're not testing for encoding here so lets make the actual output line breaks == expected line breaks
			text := strings.Replace(e.Serialize(), "\r\n", "\n", -1)

			assert.Equal(t, tc.output, text)
			assert.Equal(t, nil, err)
		})
	}
}

func TestSetAllDay(t *testing.T) {
	date, _ := time.Parse(time.RFC822, time.RFC822)

	testCases := []struct {
		name   string
		start  time.Time
		end    time.Time
		output string
	}{
		{
			name:  "test set duration - start",
			start: date,
			output: `BEGIN:VEVENT
UID:test-duration
DTSTART;VALUE=DATE:20060102
DTEND;VALUE=DATE:20060103
END:VEVENT
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEvent("test-duration")
			e.SetAllDayStartAt(date)
			e.SetAllDayEndAt(date.AddDate(0, 0, 1))

			// we're not testing for encoding here so lets make the actual output line breaks == expected line breaks
			text := strings.Replace(e.Serialize(), "\r\n", "\n", -1)

			assert.Equal(t, tc.output, text)
		})
	}
}

func TestGetLastModifiedAt(t *testing.T) {
	e := NewEvent("test-last-modified")
	lastModified := time.Unix(123456789, 0)
	e.SetLastModifiedAt(lastModified)
	got, err := e.GetLastModifiedAt()
	if err != nil {
		t.Fatalf("e.GetLastModifiedAt: %v", err)
	}

	if !got.Equal(lastModified) {
		t.Errorf("got last modified = %q, want %q", got, lastModified)
	}
}

func TestSetMailtoPrefix(t *testing.T) {
	e := NewEvent("test-set-organizer")

	e.SetOrganizer("org1@provider.com")
	if !strings.Contains(e.Serialize(), "ORGANIZER:mailto:org1@provider.com") {
		t.Errorf("expected single mailto: prefix for email org1")
	}

	e.SetOrganizer("mailto:org2@provider.com")
	if !strings.Contains(e.Serialize(), "ORGANIZER:mailto:org2@provider.com") {
		t.Errorf("expected single mailto: prefix for email org2")
	}

	e.AddAttendee("att1@provider.com")
	if !strings.Contains(e.Serialize(), "ATTENDEE:mailto:att1@provider.com") {
		t.Errorf("expected single mailto: prefix for email att1")
	}

	e.AddAttendee("mailto:att2@provider.com")
	if !strings.Contains(e.Serialize(), "ATTENDEE:mailto:att2@provider.com") {
		t.Errorf("expected single mailto: prefix for email att2")
	}
}
