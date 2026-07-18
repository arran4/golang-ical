package ics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type PropertyValueCheck struct {
	Key    string
	Values []string
}

func (c *PropertyValueCheck) Check(t *testing.T, output *BaseProperty) {
	v, ok := output.ICalParameters[c.Key]
	if !ok {
		t.Errorf("Key %s value is missing", c.Key)
		return
	}
	assert.Equal(t, c.Values, v)
}

func NewPropertyValueCheck(key string, properties ...string) *PropertyValueCheck {
	return &PropertyValueCheck{
		Key:    key,
		Values: properties,
	}
}

func TestPropertyParse(t *testing.T) {
	tests := []struct {
		Name     string
		Input    string
		Expected func(t *testing.T, output *BaseProperty, err error)
	}{
		{Name: "Normal attendee parse", Input: "ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(t *testing.T, output *BaseProperty, err error) {
			assert.NoError(t, err)
			assert.NotNil(t, output)
			assert.Equal(t, "ATTENDEE", output.IANAToken)
			assert.Equal(t, "mailto:employee-A@example.com", output.Value)
			for _, expected := range []*PropertyValueCheck{
				NewPropertyValueCheck("RSVP", "TRUE"),
			} {
				expected.Check(t, output)
			}
		}},
		{Name: "Attendee parse with quotes", Input: "ATTENDEE;RSVP=\"TRUE\";ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(t *testing.T, output *BaseProperty, err error) {
			assert.NoError(t, err)
			assert.NotNil(t, output)
			assert.Equal(t, "ATTENDEE", output.IANAToken)
			assert.Equal(t, "mailto:employee-A@example.com", output.Value)
			for _, expected := range []*PropertyValueCheck{
				NewPropertyValueCheck("RSVP", "TRUE"),
			} {
				expected.Check(t, output)
			}
		}},
		{Name: "Attendee parse with bad quotes", Input: "ATTENDEE;RSVP=T\"RUE\";ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(t *testing.T, output *BaseProperty, err error) {
			assert.Nil(t, output)
			assert.Error(t, err)
		}},
		{Name: "Attendee parse with weird escapes in quotes", Input: "ATTENDEE;CUTYPE=INDIVIDUAL;ROLE=REQ-PARTICIPANT;PARTSTAT=DECLINED;CN=xxxxxx.xxxxxxxxxx@xxxxxxxxxxx.com;X-NUM-GUESTS=0;X-RESPONSE-COMMENT=\"Abgelehnt\\, weil ich au&szlig\\;er Haus bin\":mailto:xxxxxx.xxxxxxxxxx@xxxxxxxxxxx.com", Expected: func(t *testing.T, output *BaseProperty, err error) {
			assert.NotNil(t, output)
			assert.NoError(t, err)
			assert.Equal(t, "ATTENDEE", output.IANAToken)
			assert.Equal(t, "mailto:xxxxxx.xxxxxxxxxx@xxxxxxxxxxx.com", output.Value)
			for _, expected := range []*PropertyValueCheck{
				NewPropertyValueCheck("CUTYPE", "INDIVIDUAL"),
				NewPropertyValueCheck("ROLE", "REQ-PARTICIPANT"),
				NewPropertyValueCheck("PARTSTAT", "DECLINED"),
				NewPropertyValueCheck("CN", "xxxxxx.xxxxxxxxxx@xxxxxxxxxxx.com"),
				NewPropertyValueCheck("X-NUM-GUESTS", "0"),
				NewPropertyValueCheck("X-RESPONSE-COMMENT", "Abgelehnt, weil ich au&szlig;er Haus bin"),
			} {
				expected.Check(t, output)
			}
		}},
		{Name: "Attendee parse with weird escapes in quotes short", Input: "ATTENDEE;X-RESPONSE-COMMENT=\"Abgelehnt\\, weil ich au&szlig\\;er Haus bin\":mailto:xxxxxx.xxxxxxxxxx@xxxxxxxxxxx.com\n", Expected: func(t *testing.T, output *BaseProperty, err error) {
			assert.NotNil(t, output)
			assert.NoError(t, err)
			assert.Equal(t, "ATTENDEE", output.IANAToken)
			assert.Equal(t, "mailto:xxxxxx.xxxxxxxxxx@xxxxxxxxxxx.com", output.Value)
			for _, expected := range []*PropertyValueCheck{
				NewPropertyValueCheck("X-RESPONSE-COMMENT", "Abgelehnt, weil ich au&szlig;er Haus bin"),
			} {
				expected.Check(t, output)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			v, err := parseProperty(ContentLine(test.Input))
			test.Expected(t, v, err)
		})
	}
}

func Test_parsePropertyParamValue(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		position    int
		match       string
		newposition int
		wantErr     bool
	}{
		{
			name:        "Basic sentence",
			input:       "basic sentence",
			position:    0,
			match:       "basic sentence",
			newposition: len("basic sentence"),
			wantErr:     false,
		},
		{
			name:        "Basic quoted sentence",
			input:       "\"basic sentence\"",
			position:    0,
			match:       "basic sentence",
			newposition: len("basic sentence\"\""),
			wantErr:     false,
		},
		{
			name:        "Basic sentence with terminal ,",
			input:       "basic sentence,",
			position:    0,
			match:       "basic sentence",
			newposition: len("basic sentence"),
			wantErr:     false,
		},
		{
			name:        "Basic sentence with terminal ;",
			input:       "basic sentence;",
			position:    0,
			match:       "basic sentence",
			newposition: len("basic sentence"),
			wantErr:     false,
		},
		{
			name:        "Basic sentence with terminal :",
			input:       "basic sentence:",
			position:    0,
			match:       "basic sentence",
			newposition: len("basic sentence"),
			wantErr:     false,
		},
		{
			name:        "Basic quoted sentence with terminals internal ;:,",
			input:       "\"basic sentence;:,\"",
			position:    0,
			match:       "basic sentence;:,",
			newposition: len("basic sentence;:,\"\""),
			wantErr:     false,
		},
		{
			name:        "Basic quoted sentence with escaped terminals internal ;:,",
			input:       "\"basic sentence\\;\\:\\,\"",
			position:    0,
			match:       "basic sentence;:,",
			newposition: len("basic sentence\\;\\:\\,\"\""),
			wantErr:     false,
		},
		{
			name:        "Basic quoted sentence with escaped quote",
			input:       "\"basic \\\"sentence\"",
			position:    0,
			match:       "basic \"sentence",
			newposition: len("basic sentence\\\"\"\""),
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parsePropertyParamValue(tt.input, tt.position)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePropertyParamValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.match {
				t.Errorf("parsePropertyParamValue() got = %v, want %v", got, tt.match)
			}
			if got1 != tt.newposition {
				t.Errorf("parsePropertyParamValue() got1 = %v, want %v", got1, tt.newposition)
			}
		})
	}
}

func Test_trimUT8StringUpTo(t *testing.T) {
	tests := []struct {
		name      string
		maxLength int
		s         string
		want      string
	}{
		{
			name:      "simply break at spaces",
			s:         "simply break at spaces",
			maxLength: 14,
			want:      "simply break",
		},
		{
			name:      "(Don't) Break after punctuation 1", // See if we can change this.
			s:         "hi.are.",
			maxLength: len("hi.are"),
			want:      "hi.are",
		},
		{
			name:      "Break after punctuation 2",
			s:         "Hi how are you?",
			maxLength: len("Hi how are you"),
			want:      "Hi how are",
		},
		{
			name:      "HTML opening tag breaking",
			s:         "I want a custom linkout for Thunderbird.<br>This is the Github<a href=\"https://github.com/arran4/golang-ical/issues/97\">Issue</a>.",
			maxLength: len("I want a custom linkout for Thunderbird.<br>This is the Github<"),
			want:      "I want a custom linkout for Thunderbird.<br>This is the Github",
		},
		{
			name:      "HTML closing tag breaking",
			s:         "I want a custom linkout for Thunderbird.<br>This is the Github<a href=\"https://github.com/arran4/golang-ical/issues/97\">Issue</a>.",
			maxLength: len("I want a custom linkout for Thunderbird.<br>") + 1,
			want:      "I want a custom linkout for Thunderbird.<br>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, trimUT8StringUpTo(tt.maxLength, tt.s), "trimUT8StringUpTo(%v, %v)", tt.maxLength, tt.s)
		})
	}
}

func TestFixValueStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello;world", "hello\\;world"},
		{"path\\to:file", "path\\\\to\\:file"},
		{"name:\"value\"", "name\\:\\\"value\\\""},
		{"key,value", "key\\,value"},
		{";:\\\",", "\\;\\:\\\\\\\"\\,"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeValueString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestToText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// unchanged escaping of the existing specials
		{"backslash", `a\b`, `a\\b`},
		{"semicolon", "a;b", `a\;b`},
		{"comma", "a,b", `a\,b`},
		{"lf", "a\nb", `a\nb`},
		// a literal backslash-n must not be double-escaped, only the backslash
		{"literal escape stays single", `a\nb`, `a\\nb`},
		// line-break normalisation: CR, CRLF and LF all become a single \n
		{"cr", "a\rb", `a\nb`},
		{"crlf", "line1\r\nline2", `line1\nline2`},
		{"double crlf", "a\r\n\r\nb", `a\n\nb`},
		{"lf then cr", "a\n\rb", `a\n\nb`},
		// HTAB is the one control RFC 5545 permits in TEXT
		{"htab kept", "a\tb", "a\tb"},
		// every other control character has no escape and is dropped
		{"nul", "a\x00b", "ab"},
		{"bel", "a\x07b", "ab"},
		{"vertical tab", "a\x0bb", "ab"},
		{"form feed", "a\x0cb", "ab"},
		{"escape", "a\x1bb", "ab"},
		{"del", "a\x7fb", "ab"},
		{"mixed", "x,y;z\\w\ne\rf", `x\,y\;z\\w\ne\nf`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToText(tt.input))
		})
	}
}

// TestSerializeNoRawControlCharacters checks that no TEXT value can put a
// disallowed raw control byte into a serialized content line, across the whole
// control range 0x00-0x1F and 0x7F (RFC 5545 §3.3.11 CONTROL, HTAB excepted).
func TestSerializeNoRawControlCharacters(t *testing.T) {
	for b := 0; b <= 0x7F; b++ {
		if b > 0x1F && b != 0x7F {
			continue // not a control character
		}
		cal := NewCalendar()
		event := cal.AddEvent("uid")
		event.SetProperty(ComponentPropertySummary, "A"+string(rune(b))+"B")
		serialized := cal.Serialize(WithNewLineWindows)
		// with the CRLF delimiter stripped, nothing left in a content line may
		// be a disallowed control byte
		for _, line := range strings.Split(serialized, "\r\n") {
			for i := 0; i < len(line); i++ {
				c := line[i]
				if c == '\t' {
					continue // HTAB is allowed
				}
				if c < 0x20 || c == 0x7F {
					t.Fatalf("input control 0x%02x produced raw control 0x%02x in content line %q", b, c, line)
				}
			}
		}
	}
}
