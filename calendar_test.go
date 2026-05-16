package ics

import (
	"bytes"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata/*
	TestData embed.FS
)

func captureMalformedPropertyParser(lines *[]string, next PropertyParser) PropertyParser {
	return func(rawLine ContentLine) (*BaseProperty, error) {
		line, err := parseProperty(rawLine)
		if err == nil {
			return line, nil
		}
		if lines != nil {
			*lines = append(*lines, string(rawLine))
		}
		if next == nil {
			return nil, nil
		}
		return next(rawLine)
	}
}

func TestTimeParsing(t *testing.T) {
	calFile, err := TestData.Open("testdata/timeparsing.ics")
	if err != nil {
		t.Errorf("read file: %v", err)
	}
	cal, err := ParseCalendar(calFile)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("parse calendar: %v", err)
	}

	cphLoc, locErr := time.LoadLocation("Europe/Copenhagen")
	if locErr != nil {
		t.Errorf("could not load location")
	}

	var tests = []struct {
		name        string
		uid         string
		start       time.Time
		end         time.Time
		allDayStart time.Time
		allDayEnd   time.Time
	}{
		{
			"FORM 1",
			"be7c9690-d42a-40ef-b82f-1634dc5033b4",
			time.Date(1998, 1, 18, 23, 0, 0, 0, time.Local),
			time.Date(1998, 1, 19, 23, 0, 0, 0, time.Local),
			time.Date(1998, 1, 18, 0, 0, 0, 0, time.Local),
			time.Date(1998, 1, 19, 0, 0, 0, 0, time.Local),
		},
		{
			"FORM 2",
			"53634aed-1b7d-4d85-aa38-ede76a2e4fe3",
			time.Date(2022, 1, 22, 17, 0, 0, 0, time.UTC),
			time.Date(2022, 1, 22, 20, 0, 0, 0, time.UTC),
			time.Date(2022, 1, 22, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 1, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			"FORM 3",
			"269cf715-4e14-4a10-8753-f2feeb9d060e",
			time.Date(2021, 12, 7, 14, 0, 0, 0, cphLoc),
			time.Date(2021, 12, 7, 15, 0, 0, 0, cphLoc),
			time.Date(2021, 12, 7, 0, 0, 0, 0, cphLoc),
			time.Date(2021, 12, 7, 0, 0, 0, 0, cphLoc),
		},
		{
			"Unknown local date, with 'VALUE'",
			"fb54680e-7f69-46d3-9632-00aed2469f7b",
			time.Date(2021, 6, 27, 0, 0, 0, 0, time.Local),
			time.Date(2021, 6, 28, 0, 0, 0, 0, time.Local),
			time.Date(2021, 6, 27, 0, 0, 0, 0, time.Local),
			time.Date(2021, 6, 28, 0, 0, 0, 0, time.Local),
		},
		{
			"Unknown UTC date",
			"62475ad0-a76c-4fab-8e68-f99209afcca6",
			time.Date(2021, 5, 27, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 5, 28, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 5, 27, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 5, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	assertTime := func(evtUid string, exp time.Time, timeFunc func() (given time.Time, err error)) {
		given, err := timeFunc()
		if err == nil {
			if !exp.Equal(given) {
				t.Errorf("no match on '%s', expected=%v != given=%v", evtUid, exp, given)
			}
		} else {
			t.Errorf("get time on uid '%s', %v", evtUid, err)
		}
	}
	eventMap := map[string]*VEvent{}
	for _, e := range cal.Events() {
		eventMap[e.Id()] = e
	}

	for _, tt := range tests {
		t.Run(tt.uid, func(t *testing.T) {
			evt, ok := eventMap[tt.uid]
			if !ok {
				t.Errorf("Test %#v, event UID not found, %s", tt.name, tt.uid)
				return
			}

			assertTime(tt.uid, tt.start, evt.GetStartAt)
			assertTime(tt.uid, tt.end, evt.GetEndAt)
			assertTime(tt.uid, tt.allDayStart, evt.GetAllDayStartAt)
			assertTime(tt.uid, tt.allDayEnd, evt.GetAllDayEndAt)
		})
	}
}

func TestCalendarStream(t *testing.T) {
	i := `
ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:
 mailto:employee-A@example.com
DESCRIPTION:Project XYZ Review Meeting
CATEGORIES:MEETING
CLASS:PUBLIC
`
	expected := []ContentLine{
		ContentLine("ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com"),
		ContentLine("DESCRIPTION:Project XYZ Review Meeting"),
		ContentLine("CATEGORIES:MEETING"),
		ContentLine("CLASS:PUBLIC"),
	}
	c := NewCalendarStream(strings.NewReader(i))
	cont := true
	for i := 0; cont; i++ {
		l, _, err := c.ReadLine()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				cont = false
			default:
				t.Logf("Unknown error; %v", err)
				t.Fail()
				return
			}
		}
		if l == nil {
			if errors.Is(err, io.EOF) && i == len(expected) {
				cont = false
			} else {
				t.Logf("Nil response...")
				t.Fail()
				return
			}
		}
		if i < len(expected) {
			if string(*l) != string(expected[i]) {
				t.Logf("Got %s expected %s", string(*l), string(expected[i]))
				t.Fail()
			}
		} else if l != nil {
			t.Logf("Larger than expected")
			t.Fail()
			return
		}
	}
}

func TestRfc5545Sec4Examples(t *testing.T) {
	rnReplace := regexp.MustCompile("\r?\n")

	err := fs.WalkDir(TestData, "testdata/rfc5545sec4", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		inputBytes, err := fs.ReadFile(TestData, path)
		if err != nil {
			return err
		}

		input := rnReplace.ReplaceAllString(string(inputBytes), "\r\n")
		structure, err := ParseCalendar(strings.NewReader(input))
		if err != nil && !errors.Is(err, io.EOF) {
			assert.Nil(t, err, path)
			// This should fail as the sample data doesn't conform to https://tools.ietf.org/html/rfc5545#page-45
			// Probably due to RFC width guides
			assert.NotNil(t, structure)

			output := structure.Serialize()
			assert.NotEqual(t, input, output)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("cannot read test directory: %v", err)
	}
}

func TestLineFolding(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:  "fold lines at nearest space",
			input: "some really long line with spaces to fold on and the line should fold",
			output: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:some really long line with spaces to fold on and the line
  should fold
END:VCALENDAR
`,
		},
		{
			name:  "fold lines if no space",
			input: "somereallylonglinewithnospacestofoldonandthelineshouldfoldtothenextline",
			output: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:somereallylonglinewithnospacestofoldonandthelineshouldfoldtothe
 nextline
END:VCALENDAR
`,
		},
		{
			name:  "fold lines at nearest space",
			input: "some really long line with spaces howeverthelastpartofthelineisactuallytoolongtofitonsowehavetofoldpartwaythrough",
			output: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:some really long line with spaces
  howeverthelastpartofthelineisactuallytoolongtofitonsowehavetofoldpartwayt
 hrough
END:VCALENDAR
`,
		},
		{
			name:  "75 chars line should not fold",
			input: " this line is exactly 75 characters long with the property name",
			output: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//arran4//Golang ICS Library
DESCRIPTION: this line is exactly 75 characters long with the property name
END:VCALENDAR
`,
		},
		{
			name: "runes should not be split",
			// the 75 bytes mark is in the middle of a rune
			input: "éé界世界世界世界世界世界世界世界世界世界世界世界世界",
			output: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:éé界世界世界世界世界世界世界世界世界世界
 世界世界世界
END:VCALENDAR
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCalendar()
			c.SetDescription(tc.input)
			// we're not testing for encoding here so lets make the actual output line breaks == expected line breaks
			text := strings.ReplaceAll(c.Serialize(), "\r\n", "\n")

			assert.Equal(t, tc.output, text)
			assert.True(t, utf8.ValidString(text), "Serialized .ics calendar isn't valid UTF-8 string")
		})
	}
}

func TestParseCalendar(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		output       string
		parseOptions []any
	}{
		{
			name: "test custom fields in calendar",
			input: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
END:VCALENDAR
`,
			output: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
END:VCALENDAR
`,
		},
		{
			name: "test multiline description - multiple custom fields suppress",
			input: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
BEGIN:VEVENT
DESCRIPTION:blablablablablablablablablablablablablablablabl
	testtesttest
CLASS:PUBLIC
END:VEVENT
END:VCALENDAR
`,
			output: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
BEGIN:VEVENT
DESCRIPTION:blablablablablablablablablablablablablablablabltesttesttest
CLASS:PUBLIC
END:VEVENT
END:VCALENDAR
`,
		},
		{
			name: "test semicolon in attendee property parameter",
			input: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
BEGIN:VEVENT
ATTENDEE;CN=Test\;User:mailto:user@example.com
CLASS:PUBLIC
END:VEVENT
END:VCALENDAR
`,
			output: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
BEGIN:VEVENT
ATTENDEE;CN=Test\;User:mailto:user@example.com
CLASS:PUBLIC
END:VEVENT
END:VCALENDAR
`,
		},
		{
			name: "test RRULE escaping",
			input: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
BEGIN:VEVENT
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=SU
CLASS:PUBLIC
END:VEVENT
END:VCALENDAR
`,
			output: `BEGIN:VCALENDAR
VERSION:2.0
X-CUSTOM-FIELD:test
PRODID:-//arran4//Golang ICS Library
DESCRIPTION:test
BEGIN:VEVENT
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=SU
CLASS:PUBLIC
END:VEVENT
END:VCALENDAR
`,
		},
		{
			name: "test properties after components",
			input: `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VTIMEZONE
TZID:Europe/London
BEGIN:STANDARD
TZOFFSETFROM:+0100
TZOFFSETTO:+0000
DTSTART:19701025T020000
END:STANDARD
END:VTIMEZONE
TIMEZONE-ID:VT
X-WR-TIMEZONE:VT
BEGIN:VEVENT
DTSTART:20230101T120000Z
SUMMARY:Test Event
END:VEVENT
END:VCALENDAR
`,
			parseOptions: []any{WithUnknownPropertyHandler(AcceptUnknownPropertyHandler)},
			output: `BEGIN:VCALENDAR
VERSION:2.0
TIMEZONE-ID:VT
X-WR-TIMEZONE:VT
BEGIN:VTIMEZONE
TZID:Europe/London
BEGIN:STANDARD
TZOFFSETFROM:+0100
TZOFFSETTO:+0000
DTSTART:19701025T020000
END:STANDARD
END:VTIMEZONE
BEGIN:VEVENT
DTSTART:20230101T120000Z
SUMMARY:Test Event
END:VEVENT
END:VCALENDAR
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c *Calendar
			var err error
			if len(tc.parseOptions) > 0 {
				c, err = ParseCalendarWithOptions(strings.NewReader(tc.input), tc.parseOptions...)
			} else {
				c, err = ParseCalendar(strings.NewReader(tc.input))
			}
			if !assert.NoError(t, err) {
				return
			}

			// we're not testing for encoding here so lets make the actual output line breaks == expected line breaks
			text := strings.ReplaceAll(c.Serialize(), "\r\n", "\n")
			if !assert.Equal(t, tc.output, text) {
				return
			}
		})
	}
}

func TestIssue52(t *testing.T) {
	err := fs.WalkDir(TestData, "testdata/issue52", func(path string, info fs.DirEntry, _ error) error {
		if info.IsDir() {
			return nil
		}
		_, fn := filepath.Split(path)
		t.Run(fn, func(t *testing.T) {
			f, err := TestData.Open(path)
			if err != nil && errors.Is(err, io.EOF) {
				t.Fatalf("Error reading file: %s", err)
			}
			defer f.Close()

			if _, err := ParseCalendar(f); err != nil && !errors.Is(err, io.EOF) {
				t.Fatalf("Error parsing file: %s", err)
			}

		})
		return nil
	})

	if err != nil {
		t.Fatalf("cannot read test directory: %v", err)
	}
}

func TestIssue97(t *testing.T) {
	err := fs.WalkDir(TestData, "testdata/issue97", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".ics") && !strings.HasSuffix(d.Name(), ".ics_disabled") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			if strings.HasSuffix(d.Name(), ".ics_disabled") {
				t.Skipf("Test disabled")
			}
			b, err := TestData.ReadFile(path)
			if err != nil {
				t.Fatalf("Error reading file: %s", err)
			}
			ics, err := ParseCalendar(bytes.NewReader(b))
			if err != nil {
				t.Fatalf("Error parsing file: %s", err)
			}

			got := ics.Serialize(WithLineLength(74))
			if diff := cmp.Diff(string(b), got, cmp.Transformer("ToUnixText", func(a string) string {
				return strings.ReplaceAll(a, "\r\n", "\n")
			})); diff != "" {
				t.Errorf("ParseCalendar() mismatch (-want +got):\n%s", diff)
				t.Errorf("Complete got:\b%s", got)
			}
		})
		return nil
	})

	if err != nil {
		t.Fatalf("cannot read test directory: %v", err)
	}
}

type MockHttpClient struct {
	Response *http.Response
	Error    error
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.Response, m.Error
}

var (
	_ HttpClientLike = &MockHttpClient{}
	//go:embed "testdata/rfc5545sec4/input1.ics"
	input1TestData []byte
)

func TestIssue77(t *testing.T) {
	url := "https://proseconsult.umontpellier.fr/jsp/custom/modules/plannings/direct_cal.jsp?data=58c99062bab31d256bee14356aca3f2423c0f022cb9660eba051b2653be722c4c7f281e4e3ad06b85d3374100ac416a4dc5c094f7d1a811b903031bde802c7f50e0bd1077f9461bed8f9a32b516a3c63525f110c026ed6da86f487dd451ca812c1c60bb40b1502b6511435cf9908feb2166c54e36382c1aa3eb0ff5cb8980cdb,1"

	_, err := ParseCalendarFromUrl(url, &MockHttpClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(input1TestData)),
		},
	})

	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
}

func TestWithPropertyParser_Skip(t *testing.T) {
	// Tockify-style property with underscore in parameter name.
	// Without the parser override, this aborts the entire calendar parse.
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
X-TKF-PROMOTION-BUTTON;skip_details=false;button_text=Buy Tickets:https://example.com
END:VEVENT
END:VCALENDAR
`
	// Verify it fails without the parser override.
	_, err := ParseCalendar(strings.NewReader(input))
	assert.Error(t, err, "should fail without handler")
	assert.Contains(t, err.Error(), "missing property value for skip")

	// With the parser override set to skip, parsing succeeds.
	var skippedLines []string
	cal, err := ParseCalendarWithOptions(strings.NewReader(input),
		captureMalformedPropertyParser(&skippedLines, SkipPropertyParser),
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, cal)
	assert.Len(t, cal.Events(), 1)
	assert.Equal(t, "Test Event", cal.Events()[0].GetProperty(ComponentPropertySummary).Value)
	assert.Len(t, skippedLines, 1)
	assert.Contains(t, skippedLines[0], "X-TKF-PROMOTION-BUTTON")
}

func TestWithPropertyParser_Abort(t *testing.T) {
	// Parser override that returns an error aborts parsing with that error.
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
X-BAD;under_score=val:value
END:VEVENT
END:VCALENDAR
`
	customErr := errors.New("custom abort")
	_, err := ParseCalendarWithOptions(strings.NewReader(input),
		func(rawLine ContentLine) (*BaseProperty, error) {
			_, parseErr := parseProperty(rawLine)
			if parseErr != nil {
				return nil, customErr
			}
			return nil, nil
		},
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErr)
}

func TestWithPropertyParser_Recover(t *testing.T) {
	// Parser override that returns a replacement property.
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
X-BAD;under_score=val:the-value
END:VEVENT
END:VCALENDAR
`
	cal, err := ParseCalendarWithOptions(strings.NewReader(input),
		FallbackParser(LooseParser),
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, cal)
	events := cal.Events()
	assert.Len(t, events, 1)
	// The recovered property should be present.
	badProp := events[0].GetProperty("X-BAD")
	if assert.NotNil(t, badProp) {
		assert.Equal(t, "the-value", badProp.Value)
	}
}

func TestWithPropertyParser_FallbackChain(t *testing.T) {
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
X-BAD;under_score=val:the-value
END:VEVENT
END:VCALENDAR
`
	var firstFallbackCalls int
	parser := FallbackParser(
		func(rawLine ContentLine) (*BaseProperty, error) {
			firstFallbackCalls++
			return nil, nil
		},
		LooseParser,
	)
	cal, err := ParseCalendarWithOptions(strings.NewReader(input), parser)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, firstFallbackCalls)
	badProp := cal.Events()[0].GetProperty("X-BAD")
	if assert.NotNil(t, badProp) {
		assert.Equal(t, "the-value", badProp.Value)
	}
}

func TestWithPropertyParser_CalendarLevel(t *testing.T) {
	// Malformed property at the calendar level (outside any component).
	input := `BEGIN:VCALENDAR
VERSION:2.0
X-BAD-CAL;under_score=val:cal-value
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
END:VEVENT
END:VCALENDAR
`
	var skippedLines []string
	cal, err := ParseCalendarWithOptions(strings.NewReader(input),
		captureMalformedPropertyParser(&skippedLines, SkipPropertyParser),
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, cal)
	assert.Len(t, skippedLines, 1)
	assert.Contains(t, skippedLines[0], "X-BAD-CAL")
	// Valid properties should still be parsed.
	assert.Len(t, cal.Events(), 1)
}

func TestWithPropertyParser_NoHandler(t *testing.T) {
	// Without the parser override, malformed properties still cause errors (backward compat).
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
X-BAD;under_score=val:value
END:VEVENT
END:VCALENDAR
`
	_, err := ParseCalendar(strings.NewReader(input))
	assert.Error(t, err)
}

func TestParseCalendar_MalformedErrorIncludesLine(t *testing.T) {
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
X-BAD;under_score=val:value
END:VEVENT
END:VCALENDAR
`

	_, err := ParseCalendar(strings.NewReader(input))
	if !assert.Error(t, err) {
		return
	}
	var malformed *MalformedError
	if assert.True(t, errors.As(err, &malformed)) {
		assert.Equal(t, 7, malformed.Line)
		assert.False(t, malformed.HasChar)
		assert.ErrorIs(t, err, ErrMissingPropertyValue)
	}
}

func TestParseCalendar_MalformedErrorFirstLine(t *testing.T) {
	input := `BOGUS:VCALENDAR
VERSION:2.0
END:VCALENDAR
`

	_, err := ParseCalendar(strings.NewReader(input))
	if !assert.Error(t, err) {
		return
	}
	var malformed *MalformedError
	if assert.True(t, errors.As(err, &malformed)) {
		assert.Equal(t, 1, malformed.Line)
		assert.False(t, malformed.HasChar)
		assert.ErrorIs(t, err, ErrExpectedBegin)
	}
}

func TestParseCalendar_MalformedErrorNestedComponentLine(t *testing.T) {
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
BEGIN:VALARM
TRIGGER:-PT15M
ACTION:DISPLAY
X-ALARM-BAD;under_score=val:alarm-value
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
END:VCALENDAR
`

	_, err := ParseCalendar(strings.NewReader(input))
	if !assert.Error(t, err) {
		return
	}
	var malformed *MalformedError
	if assert.True(t, errors.As(err, &malformed)) {
		assert.Equal(t, 10, malformed.Line)
		assert.False(t, malformed.HasChar)
		assert.ErrorIs(t, err, ErrMissingPropertyValue)
	}
}

func TestWithPropertyParser_NestedComponent(t *testing.T) {
	// Parser override should propagate into nested components (e.g. VALARM inside VEVENT).
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
BEGIN:VALARM
TRIGGER:-PT15M
ACTION:DISPLAY
X-ALARM-BAD;under_score=val:alarm-value
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
END:VCALENDAR
`
	var skippedLines []string
	cal, err := ParseCalendarWithOptions(strings.NewReader(input),
		captureMalformedPropertyParser(&skippedLines, SkipPropertyParser),
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, cal)
	assert.Len(t, skippedLines, 1)
	assert.Contains(t, skippedLines[0], "X-ALARM-BAD")
	events := cal.Events()
	assert.Len(t, events, 1)
	alarms := events[0].Alarms()
	assert.Len(t, alarms, 1)
	// The valid alarm properties should still be there.
	assert.NotNil(t, alarms[0].GetProperty(ComponentPropertyTrigger))
	assert.NotNil(t, alarms[0].GetProperty(ComponentPropertyAction))
}

func TestWithPropertyParser_CalendarLevelRecover(t *testing.T) {
	input := `BEGIN:VCALENDAR
VERSION:2.0
X-BAD-CAL;under_score=val:cal-value
PRODID:-//Test//Test//EN
END:VCALENDAR
`

	cal, err := ParseCalendarWithOptions(strings.NewReader(input),
		FallbackParser(LooseParser),
	)
	if !assert.NoError(t, err) {
		return
	}

	found := false
	for _, p := range cal.CalendarProperties {
		if p.IANAToken == "X-BAD-CAL" {
			found = true
			assert.Equal(t, "cal-value", p.Value)
		}
	}
	assert.True(t, found, "expected recovered calendar-level property")
}

func TestWithPropertyParser_NestedComponentRecover(t *testing.T) {
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
SUMMARY:Test Event
BEGIN:VALARM
TRIGGER:-PT15M
ACTION:DISPLAY
X-ALARM-BAD;under_score=val:alarm-value
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
END:VCALENDAR
`

	cal, err := ParseCalendarWithOptions(strings.NewReader(input),
		FallbackParser(LooseParser),
	)
	if !assert.NoError(t, err) {
		return
	}

	events := cal.Events()
	assert.Len(t, events, 1)
	alarms := events[0].Alarms()
	assert.Len(t, alarms, 1)
	recovered := alarms[0].GetProperty("X-ALARM-BAD")
	if assert.NotNil(t, recovered) {
		assert.Equal(t, "alarm-value", recovered.Value)
	}
}

func TestWithPropertyParser_MultipleConsecutive(t *testing.T) {
	// Multiple malformed properties in the same component should all be
	// passed to the parser override individually.
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
DTSTART:20240101T120000Z
X-BAD-ONE;under_score=val:value1
SUMMARY:Test Event
X-BAD-TWO;another_bad=yes:value2
X-BAD-THREE;yet_another=no:value3
DTEND:20240101T130000Z
END:VEVENT
END:VCALENDAR
`
	var skippedLines []string
	cal, err := ParseCalendarWithOptions(strings.NewReader(input),
		captureMalformedPropertyParser(&skippedLines, SkipPropertyParser),
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, cal)
	assert.Len(t, cal.Events(), 1)
	// All three malformed properties should have been skipped.
	assert.Len(t, skippedLines, 3)
	assert.Contains(t, skippedLines[0], "X-BAD-ONE")
	assert.Contains(t, skippedLines[1], "X-BAD-TWO")
	assert.Contains(t, skippedLines[2], "X-BAD-THREE")
	// Valid properties should still be present.
	ev := cal.Events()[0]
	assert.Equal(t, "Test Event", ev.GetProperty(ComponentPropertySummary).Value)
	assert.NotNil(t, ev.GetProperty(ComponentPropertyDtStart))
	assert.NotNil(t, ev.GetProperty(ComponentPropertyDtEnd))
}

func TestParseCalendarWithOptions_InvalidOptionType(t *testing.T) {
	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
END:VCALENDAR
`

	_, err := ParseCalendarWithOptions(strings.NewReader(input), 42)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, ErrInvalidOpArg)
		assert.Contains(t, err.Error(), "0")
	}
}

func TestNewCalendarWithOptions_InvalidOptionType(t *testing.T) {
	_, err := NewCalendarWithOptions(42)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, ErrInvalidOpArg)
		assert.Contains(t, err.Error(), "0")
	}
}

func TestParseCalendarFromUrl_InvalidOptionType(t *testing.T) {
	_, err := ParseCalendarFromUrl("https://example.com", 42)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, ErrInvalidOpArg)
		assert.Contains(t, err.Error(), "0")
	}
}

func TestComponentParseWithOptions_InvalidOptionType(t *testing.T) {
	_, err := ParseComponentWithOptions(nil, nil, 42)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, ErrInvalidOpArg)
		assert.Contains(t, err.Error(), "0")
	}

	_, err = GeneralParseComponentWithOptions(nil, nil, 42)
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, ErrInvalidOpArg)
		assert.Contains(t, err.Error(), "0")
	}
}

func TestSerialize_InvalidOptionType(t *testing.T) {
	_, err := parseSerializeOps([]any{42})
	if assert.Error(t, err) {
		assert.ErrorIs(t, err, ErrInvalidOpArg)
		assert.Contains(t, err.Error(), "0")
	}
}

func TestParseProperty_ExportedStrict(t *testing.T) {
	line, err := ParseProperty(ContentLine("SUMMARY:ok"))
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, line) {
		assert.Equal(t, "SUMMARY", line.IANAToken)
		assert.Equal(t, "ok", line.Value)
	}

	_, err = ParseProperty(ContentLine("X-BAD;under_score=val:value"))
	assert.Error(t, err)
}

func TestErrPropertySkippedBehavior(t *testing.T) {
	parser := func(rawLine ContentLine) (*BaseProperty, error) {
		if strings.Contains(string(rawLine), "X-SKIP") {
			return nil, fmt.Errorf("%w: intentional", ErrPropertySkipped)
		}
		return parseProperty(rawLine)
	}

	input := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
X-SKIP:1
BEGIN:VEVENT
SUMMARY:ok
X-SKIP:2
END:VEVENT
END:VCALENDAR
`

	cal, err := ParseCalendarWithOptions(strings.NewReader(input), parser)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, cal.Events(), 1)
	assert.Equal(t, "ok", cal.Events()[0].GetProperty(ComponentPropertySummary).Value)
}

func BenchmarkSerialize(b *testing.B) {
	calFile, err := TestData.Open("testdata/serialization/input2.ics")
	require.NoError(b, err)

	cal, err := ParseCalendar(calFile)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cal.Serialize()
	}
}
