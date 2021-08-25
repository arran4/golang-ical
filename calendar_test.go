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

	"github.com/stretchr/testify/assert"
)

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

//Is that what this is actually testing?
func TestCreatesUTF8String(t *testing.T) {
	c := NewCalendar()
	// Repeating a string that contains long runes (size > 1 byte)
	// 75 time to be sure that line folding is needed.
	const runesToOverflowLine = 75
	c.SetDescription(strings.Repeat("世界", runesToOverflowLine))
	text := c.Serialize()
	assert.True(t, utf8.ValidString(text), "Serialized .ics calendar isn't valid UTF-8 string")
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCalendar()
			// Repeating a string that contains long runes (size > 1 byte)
			// 75 time to be sure that line folding is needed.
			const runesToOverflowLine = 75
			assert.Greaterf(t, len("DESCRIPTION:"+tc.input), runesToOverflowLine, "Test data has issue")
			c.SetDescription(tc.input)
			// we're not testing for encoding here so lets make the actual output line breaks == expected line breaks
			text := strings.Replace(c.Serialize(), "\r\n", "\n", -1)

			assert.Equal(t, tc.output, text)
		})
	}
}
