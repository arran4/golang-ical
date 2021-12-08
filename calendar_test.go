package ics

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestTimeParsing(t *testing.T) {
	calFile, err := os.OpenFile("./testdata/timeparsing.ics", os.O_RDONLY, 0400)
	if err != nil {
		t.Errorf("read file: %v", err)
	}
	cal, err := ParseCalendar(calFile)
	if err != nil {
		t.Errorf("parse calendar: %v", err)
	}

	cphLoc, locErr := time.LoadLocation("Europe/Copenhagen")
	if locErr != nil {
		t.Errorf("could not load location")
	}

	var tests = []struct {
		uid         string
		start       time.Time
		end         time.Time
		allDayStart time.Time
		allDayEnd   time.Time
	}{
		// FORM 1
		{"be7c9690-d42a-40ef-b82f-1634dc5033b4",
			time.Date(1998, 1, 18, 23, 0, 0, 0, time.Local),
			time.Date(1998, 1, 19, 23, 0, 0, 0, time.Local),
			time.Date(1998, 1, 18, 0, 0, 0, 0, time.Local),
			time.Date(1998, 1, 19, 0, 0, 0, 0, time.Local)},
		// FORM 2
		{"53634aed-1b7d-4d85-aa38-ede76a2e4fe3",
			time.Date(2022, 1, 22, 17, 0, 0, 0, time.UTC),
			time.Date(2022, 1, 22, 20, 0, 0, 0, time.UTC),
			time.Date(2022, 1, 22, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 1, 22, 0, 0, 0, 0, time.UTC)},
		// FORM 3
		{"269cf715-4e14-4a10-8753-f2feeb9d060e",
			time.Date(2021, 12, 7, 14, 0, 0, 0, cphLoc),
			time.Date(2021, 12, 7, 15, 0, 0, 0, cphLoc),
			time.Date(2021, 12, 7, 0, 0, 0, 0, cphLoc),
			time.Date(2021, 12, 7, 0, 0, 0, 0, cphLoc)},
		// Unknown local date, with 'VALUE'
		{"fb54680e-7f69-46d3-9632-00aed2469f7b",
			time.Date(2021, 6, 27, 0, 0, 0, 0, time.Local),
			time.Date(2021, 6, 28, 0, 0, 0, 0, time.Local),
			time.Date(2021, 6, 27, 0, 0, 0, 0, time.Local),
			time.Date(2021, 6, 28, 0, 0, 0, 0, time.Local)},
		// Unknown UTC date
		{"62475ad0-a76c-4fab-8e68-f99209afcca6",
			time.Date(2021, 5, 27, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 5, 28, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 5, 27, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 5, 28, 0, 0, 0, 0, time.UTC)},
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
	evts := cal.Events()

	for _, tt := range tests {
		t.Run(tt.uid, func(t *testing.T) {
			var evt *VEvent
			for _, e := range evts {
				if strings.EqualFold(e.Id(), tt.uid) {
					evt = e
				}
			}

			if evt == nil {
				t.Errorf("event UID not found, %s", tt.uid)
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
		l, err := c.ReadLine()
		if err != nil {
			switch err {
			case io.EOF:
				cont = false
			default:
				t.Logf("Unknown error; %v", err)
				t.Fail()
				return
			}
		}
		if l == nil {
			if err == io.EOF && i == len(expected) {
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

	err := filepath.Walk("./testdata/rfc5545sec4/", func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		inputBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		input := rnReplace.ReplaceAllString(string(inputBytes), "\r\n")
		structure, err := ParseCalendar(strings.NewReader(input))
		if assert.Nil(t, err, path) {
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
			text := strings.Replace(c.Serialize(), "\r\n", "\n", -1)

			assert.Equal(t, tc.output, text)
			assert.True(t, utf8.ValidString(text), "Serialized .ics calendar isn't valid UTF-8 string")
		})
	}
}
