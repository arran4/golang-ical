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
			text := strings.ReplaceAll(e.Serialize(defaultSerializationOptions()), "\r\n", "\n")

			assert.Equal(t, tc.output, text)
			assert.Equal(t, nil, err)
		})
	}
}

func TestSetGeo(t *testing.T) {
	testCases := []struct {
		name     string
		lat      interface{}
		lng      interface{}
		expected string
	}{
		{
			name:     "float",
			lat:      1.23,
			lng:      4.56,
			expected: "1.23;4.56",
		},
		{
			name:     "int",
			lat:      1,
			lng:      2,
			expected: "1;2",
		},
		{
			name:     "string",
			lat:      "1.234",
			lng:      "4.567",
			expected: "1.234;4.567",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEvent("test-geo")

			switch lat := tc.lat.(type) {
			case float64:
				SetGeo(e, lat, tc.lng.(float64))
			case int:
				SetGeo(e, lat, tc.lng.(int))
			case string:
				SetGeo(e, lat, tc.lng.(string))
			}

			prop := e.GetProperty(ComponentPropertyGeo)
			assert.NotNil(t, prop)
			assert.Equal(t, tc.expected, prop.Value)
		})
	}

	t.Run("legacy", func(t *testing.T) {
		e := NewEvent("test-geo-legacy")
		e.SetGeo(1.23, 4.56)
		prop := e.GetProperty(ComponentPropertyGeo)
		assert.NotNil(t, prop)
		assert.Equal(t, "1.23;4.56", prop.Value)
	})
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

			text := strings.ReplaceAll(e.Serialize(defaultSerializationOptions()), "\r\n", "\n")

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
	if !strings.Contains(e.Serialize(defaultSerializationOptions()), "ORGANIZER:mailto:org1@provider.com") {
		t.Errorf("expected single mailto: prefix for email org1")
	}

	e.SetOrganizer("mailto:org2@provider.com")
	if !strings.Contains(e.Serialize(defaultSerializationOptions()), "ORGANIZER:mailto:org2@provider.com") {
		t.Errorf("expected single mailto: prefix for email org2")
	}

	e.AddAttendee("att1@provider.com")
	if !strings.Contains(e.Serialize(defaultSerializationOptions()), "ATTENDEE:mailto:att1@provider.com") {
		t.Errorf("expected single mailto: prefix for email att1")
	}

	e.AddAttendee("mailto:att2@provider.com")
	if !strings.Contains(e.Serialize(defaultSerializationOptions()), "ATTENDEE:mailto:att2@provider.com") {
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
			text := strings.ReplaceAll(e.Serialize(defaultSerializationOptions()), "\r\n", "\n")

			assert.Equal(t, tc.output, text)
		})
	}
}

func TestRemovePropertyByValue(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   string
		removed []int
		kept    []int
	}{
		{"matching attendees", "mailto:a@x", []int{1, 3}, []int{0, 2, 4}},
		{"other property value", "u", nil, []int{0, 1, 2, 3, 4}},
		{"missing value", "mailto:missing@x", nil, []int{0, 1, 2, 3, 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cal := NewCalendar()
			e := cal.AddEvent("u")
			e.AddAttendee("a@x", WithCN("First"))
			e.AddAttendee("b@x", WithCN("Second"))
			e.AddAttendee("a@x", WithCN("Duplicate"))
			e.SetSummary("mailto:a@x")

			var wantRemoved, wantKept []IANAProperty
			for _, i := range tc.removed {
				wantRemoved = append(wantRemoved, e.Properties[i])
			}
			for _, i := range tc.kept {
				wantKept = append(wantKept, e.Properties[i])
			}

			removed := e.RemovePropertyByValue(ComponentPropertyAttendee, tc.value)

			assert.Equal(t, wantRemoved, removed)
			assert.Equal(t, wantKept, e.Properties)
			parsed, err := ParseCalendar(strings.NewReader(cal.Serialize()))
			if assert.NoError(t, err) && assert.Len(t, parsed.Events(), 1) {
				assert.Equal(t, cal.Serialize(), parsed.Serialize())
			}
		})
	}
}

func TestRemovePropertyByFunc(t *testing.T) {
	for _, tc := range []struct {
		name     string
		property ComponentProperty
		remove   func(IANAProperty) bool
		removed  []int
		kept     []int
		visited  []int
	}{
		{
			name: "matching parameter", property: ComponentPropertyAttendee,
			remove:  func(p IANAProperty) bool { return p.ICalParameters["CN"][0] == "Remove" },
			removed: []int{2}, kept: []int{0, 1, 3}, visited: []int{1, 2},
		},
		{
			name: "no predicate matches", property: ComponentPropertyAttendee,
			remove: func(IANAProperty) bool { return false },
			kept:   []int{0, 1, 2, 3}, visited: []int{1, 2},
		},
		{
			name: "all predicate matches", property: ComponentPropertyAttendee,
			remove:  func(IANAProperty) bool { return true },
			removed: []int{1, 2}, kept: []int{0, 3}, visited: []int{1, 2},
		},
		{
			name: "missing property", property: ComponentPropertyComment,
			remove: func(IANAProperty) bool { return true },
			kept:   []int{0, 1, 2, 3},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEvent("u")
			e.AddAttendee("a@x", WithCN("Keep"))
			e.AddAttendee("a@x", WithCN("Remove"))
			e.SetSummary("s")

			var wantRemoved, wantKept, wantVisited []IANAProperty
			for _, i := range tc.removed {
				wantRemoved = append(wantRemoved, e.Properties[i])
			}
			for _, i := range tc.kept {
				wantKept = append(wantKept, e.Properties[i])
			}
			for _, i := range tc.visited {
				wantVisited = append(wantVisited, e.Properties[i])
			}

			var visited []IANAProperty
			removed := e.RemovePropertyByFunc(tc.property, func(p IANAProperty) bool {
				visited = append(visited, p)
				if !assert.Equal(t, string(tc.property), p.IANAToken) {
					return false
				}
				return tc.remove(p)
			})

			assert.Equal(t, wantRemoved, removed)
			assert.Equal(t, wantKept, e.Properties)
			assert.Equal(t, wantVisited, visited)
		})
	}
}
