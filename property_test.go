package ics

import (
	"fmt"
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

func TestParamValueEscapersDropControlCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		unquoted string
		quoted   string
	}{
		{"htab is WSP and stays", "a\tb", "a\tb", "\"a\tb\""},
		{"nul", "a\x00b", "ab", `"ab"`},
		{"bel", "a\ab", "ab", `"ab"`},
		{"lf", "a\nb", "ab", `"ab"`},
		{"cr", "a\rb", "ab", `"ab"`},
		{"crlf", "a\r\nb", "ab", `"ab"`},
		{"vertical tab", "a\vb", "ab", `"ab"`},
		{"form feed", "a\fb", "ab", `"ab"`},
		{"escape", "a\x1bb", "ab", `"ab"`},
		{"del", "a\x7fb", "ab", `"ab"`},
		{"non-ascii is not a control", "a\u0080b", "a\u0080b", "\"a\u0080b\""},
		{"specials are still escaped", `a,b;c:d"e\f`, `a\,b\;c\:d\"e\\f`, `"a,b;c:d\"e\\f"`},
		{"control between specials", "a\x00,b", `a\,b`, `"a,b"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.unquoted, escapeValueString(tt.input))
			assert.Equal(t, tt.quoted, quotedValueString(tt.input))
		})
	}
}

func serializeParamValue(t *testing.T, key, value string) string {
	t.Helper()
	cal := NewCalendar()
	event := cal.AddEvent("uid")
	event.SetProperty(ComponentProperty("X-TEST"), "v", &KeyValues{Key: key, Value: []string{value}})
	return cal.Serialize(WithNewLineWindows)
}

func parsedParamValue(t *testing.T, serialized, key string) (string, bool) {
	t.Helper()
	cal, err := ParseCalendar(strings.NewReader(serialized))
	if !assert.NoError(t, err, "serialized output must parse back: %q", serialized) {
		return "", false
	}
	for _, component := range cal.Components {
		for _, property := range component.UnknownPropertiesIANAProperties() {
			if property.IANAToken != "X-TEST" {
				continue
			}
			if values, ok := property.ICalParameters[key]; ok && len(values) > 0 {
				return values[0], true
			}
		}
	}
	return "", false
}

func assertNoRawControlCharacter(t *testing.T, serialized string) {
	t.Helper()
	for _, line := range strings.Split(serialized, string(WithNewLineWindows)) {
		for i := 0; i < len(line); i++ {
			if c := line[i]; (c < 0x20 || c == 0x7F) && c != '\t' {
				t.Fatalf("raw control 0x%02x in content line %q", c, line)
			}
		}
	}
}

// A param-value carries no escape mechanism, so serialization has to drop every
// CONTROL character rather than emit one parsePropertyParamValue rejects.
func TestSerializeParamValueNoRawControlCharacters(t *testing.T) {
	// CN is written as paramtext, ALTREP as a quoted-string.
	for _, key := range []string{"CN", "ALTREP"} {
		for b := 0x00; b <= 0x7F; b++ {
			if b > 0x1F && b != 0x7F {
				continue
			}
			t.Run(fmt.Sprintf("%s/0x%02x", key, b), func(t *testing.T) {
				serialized := serializeParamValue(t, key, "a"+string(rune(b))+"b")
				assertNoRawControlCharacter(t, serialized)

				want := "ab"
				if b == '\t' {
					want = "a\tb"
				}
				got, ok := parsedParamValue(t, serialized, key)
				assert.True(t, ok, "%s must survive the round trip, got %q", key, serialized)
				assert.Equal(t, want, got)
			})
		}
	}
}

// The strip must leave everything SAFE-CHAR and QSAFE-CHAR permit alone.
func TestSerializeParamValuePrintableUnchanged(t *testing.T) {
	for _, key := range []string{"CN", "ALTREP"} {
		for b := 0x20; b <= 0x7E; b++ {
			t.Run(fmt.Sprintf("%s/0x%02x", key, b), func(t *testing.T) {
				want := "a" + string(rune(b)) + "b"
				got, ok := parsedParamValue(t, serializeParamValue(t, key, want), key)
				assert.True(t, ok, "%s must survive the round trip", key)
				assert.Equal(t, want, got)
			})
		}
	}
}

func TestSerializeParamValueControlAcrossFold(t *testing.T) {
	for _, key := range []string{"CN", "ALTREP"} {
		t.Run(key, func(t *testing.T) {
			want := strings.Repeat("x", 70) + strings.Repeat("y", 20)
			serialized := serializeParamValue(t, key, strings.Repeat("x", 70)+"\v"+strings.Repeat("y", 20))
			assert.Contains(t, serialized, string(WithNewLineWindows)+" ", "value must be long enough to fold")
			assertNoRawControlCharacter(t, serialized)

			got, ok := parsedParamValue(t, serialized, key)
			assert.True(t, ok, "%s must survive the round trip", key)
			assert.Equal(t, want, got)
		})
	}
}
