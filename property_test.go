package ics

import (
	"testing"
	"time"

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
			v, err := ParseProperty(ContentLine(test.Input))
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

func TestParseDurations(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []Duration
		hasError bool
	}{
		{
			name:  "Valid duration with days, hours, and seconds",
			value: "P15DT5H0M20S",
			expected: []Duration{
				{Positive: true, Duration: 5*time.Hour + 20*time.Second, Days: 15},
			},
			hasError: false,
		},
		{
			name:  "Valid duration with weeks",
			value: "P7W",
			expected: []Duration{
				{Positive: true, Duration: 0, Days: 7 * 7}, // 7 weeks
			},
			hasError: false,
		},
		{
			name:  "Valid negative duration",
			value: "-P1DT3H",
			expected: []Duration{
				{Positive: false, Duration: 3 * time.Hour, Days: 1},
			},
			hasError: false,
		},
		{
			name:     "Invalid duration missing 'P'",
			value:    "15DT5H0M20S",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Invalid input format with random string",
			value:    "INVALID",
			expected: nil,
			hasError: true,
		},
		{
			name:  "Multiple durations in comma-separated list",
			value: "P1DT5H,P2DT3H",
			expected: []Duration{
				{Positive: true, Duration: 5 * time.Hour, Days: 1},
				{Positive: true, Duration: 3 * time.Hour, Days: 2},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := IANAProperty{BaseProperty{Value: tt.value}}
			durations, err := prop.ParseDurations()

			if (err != nil) != tt.hasError {
				t.Fatalf("expected error: %v, got: %v", tt.hasError, err)
			}

			if !tt.hasError && !equalDurations(durations, tt.expected) {
				t.Errorf("expected durations: %v, got: %v", tt.expected, durations)
			}
		})
	}
}

// Helper function to compare two slices of Duration
func equalDurations(a, b []Duration) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
