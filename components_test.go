package ics

import (
	"errors"
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
			text := strings.ReplaceAll(e.Serialize(), "\r\n", "\n")

			assert.Equal(t, tc.output, text)
			assert.Equal(t, nil, err)
		})
	}
}

func TestSetAllDay(t *testing.T) {
	date, _ := time.Parse(time.RFC822, time.RFC822)

	testCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		duration time.Duration
		output   string
	}{
		{
			name:  "test set all day - start",
			start: date,
			output: `BEGIN:VEVENT
UID:test-allday
DTSTART;VALUE=DATE:20060102
END:VEVENT
`,
		},
		{
			name: "test set all day - end",
			end:  date,
			output: `BEGIN:VEVENT
UID:test-allday
DTEND;VALUE=DATE:20060102
END:VEVENT
`,
		},
		{
			name:     "test set all day - duration",
			start:    date,
			duration: time.Hour * 24,
			output: `BEGIN:VEVENT
UID:test-allday
DTSTART;VALUE=DATE:20060102
DTEND;VALUE=DATE:20060103
END:VEVENT
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEvent("test-allday")
			if !tc.start.IsZero() {
				e.SetAllDayStartAt(tc.start)
			}
			if !tc.end.IsZero() {
				e.SetAllDayEndAt(tc.end)
			}
			if tc.duration != 0 {
				err := e.SetDuration(tc.duration)
				assert.NoError(t, err)
			}

			text := strings.ReplaceAll(e.Serialize(), "\r\n", "\n")

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

func TestRemoveProperty(t *testing.T) {
	testCases := []struct {
		name   string
		output string
	}{
		{
			name: "test RemoveProperty - start",
			output: `BEGIN:VTODO
UID:test-removeproperty
X-TEST:42
END:VTODO
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewTodo("test-removeproperty")
			e.AddProperty("X-TEST", "42")
			e.AddProperty("X-TESTREMOVE", "FOO")
			e.AddProperty("X-TESTREMOVE", "BAR")
			e.RemoveProperty("X-TESTREMOVE")

			// adjust to expected linebreaks, since we're not testing the encoding
			text := strings.ReplaceAll(e.Serialize(), "\r\n", "\n")

			assert.Equal(t, tc.output, text)
		})
	}
}

// Helper function to create a *time.Time from a string
func MustNewTime(value string) *time.Time {
	t, err := time.ParseInLocation(time.RFC3339, value, time.UTC)
	if err != nil {
		return nil
	}
	return &t
}

func TestIsDuring(t *testing.T) {
	tests := []struct {
		name           string
		startTime      *time.Time
		endTime        *time.Time
		duration       string
		pointInTime    time.Time
		expectedResult bool
		expectedError  error
		allDayStart    bool
		allDayEnd      bool
	}{
		{
			name:           "Valid start and end time",
			startTime:      MustNewTime("2024-10-15T09:00:00Z"),
			endTime:        MustNewTime("2024-10-15T17:00:00Z"),
			pointInTime:    time.Date(2024, 10, 15, 10, 0, 0, 0, time.UTC),
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "Valid start time, no end, duration",
			startTime:      MustNewTime("2024-10-15T09:00:00Z"),
			duration:       "P2H",
			pointInTime:    time.Date(2024, 10, 15, 11, 0, 0, 0, time.UTC),
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "No start or end time",
			pointInTime:    time.Date(2024, 10, 15, 10, 0, 0, 0, time.UTC),
			expectedResult: false,
			expectedError:  ErrStartAndEndDateNotDefined,
		},
		{
			name:           "All-day event",
			startTime:      MustNewTime("2024-10-15T00:00:00Z"),
			endTime:        MustNewTime("2024-10-15T23:59:59Z"),
			pointInTime:    time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC),
			expectedResult: true,
			expectedError:  nil,
			allDayStart:    true,
			allDayEnd:      true,
		},
		{
			name:           "Point outside event duration",
			startTime:      MustNewTime("2024-10-15T09:00:00Z"),
			endTime:        MustNewTime("2024-10-15T17:00:00Z"),
			pointInTime:    time.Date(2024, 10, 15, 18, 0, 0, 0, time.UTC),
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "All-day start with valid end time",
			startTime:      MustNewTime("2024-10-15T00:00:00Z"),
			endTime:        MustNewTime("2024-10-15T17:00:00Z"),
			pointInTime:    time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC),
			expectedResult: true,
			expectedError:  nil,
			allDayStart:    true,
		},
		{
			name:           "All-day end with valid start time",
			startTime:      MustNewTime("2024-10-15T09:00:00Z"),
			endTime:        MustNewTime("2024-10-15T23:59:59Z"),
			pointInTime:    time.Date(2024, 10, 15, 22, 0, 0, 0, time.UTC),
			expectedResult: true,
			expectedError:  nil,
			allDayEnd:      true,
		},
		{
			name:           "Duration 1 day, point within event",
			startTime:      MustNewTime("2024-10-15T09:00:00Z"),
			duration:       "P1D",
			pointInTime:    time.Date(2024, 10, 16, 10, 0, 0, 0, time.UTC),
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "Duration 2 hours, point after event",
			startTime:      MustNewTime("2024-10-15T09:00:00Z"),
			duration:       "P2H",
			pointInTime:    time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC),
			expectedResult: false,
			expectedError:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := &ComponentBase{}
			if tt.startTime != nil {
				if tt.allDayStart {
					cb.SetAllDayStartAt(*tt.startTime)
				} else {
					cb.SetStartAt(*tt.startTime)
				}
			}
			if tt.endTime != nil {
				if tt.allDayEnd {
					cb.SetAllDayEndAt(*tt.endTime)
				} else {
					cb.SetEndAt(*tt.endTime)
				}
			}
			if tt.duration != "" {
				err := cb.SetDurationStr(tt.duration)
				if err != nil {
					t.Fatalf("Duration parse failed: %s", err)
				}
			}
			// Call the IsDuring method
			result, err := cb.IsDuring(tt.pointInTime)

			if err != nil || tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
				}
			}

			if result != tt.expectedResult {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, result)
			}
		})
	}
}
