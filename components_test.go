package ics

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestParseICalDurationRFCExamples(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Duration
	}{
		{
			name:     "days hours minutes seconds",
			input:    "P15DT5H0M20S",
			expected: []Duration{{Sign: 0, Time: 5*time.Hour + 20*time.Second, Days: 15}},
		},
		{
			name:     "weeks",
			input:    "P7W",
			expected: []Duration{{Sign: 0, Time: 0, Days: 7 * 7}},
		},
		{
			name:     "negative duration",
			input:    "-P1DT3H",
			expected: []Duration{{Sign: -1, Time: 3 * time.Hour, Days: 1}},
		},
		{
			name:     "missing designator",
			input:    "15DT5H0M20S",
			expected: nil,
		},
		{
			name:     "random string",
			input:    "INVALID",
			expected: nil,
		},
		{
			name:     "comma separated list",
			input:    "P1DT5H,P2DT3H",
			expected: []Duration{{Sign: 0, Time: 5 * time.Hour, Days: 1}, {Sign: 0, Time: 3 * time.Hour, Days: 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected == nil {
				got, ok, err := ParseICalDuration(tt.input)
				require.NoError(t, err)
				require.False(t, ok)
				assert.Equal(t, Duration{}, got)
				return
			}

			parts := strings.Split(tt.input, ",")
			require.Len(t, parts, len(tt.expected))
			for i, part := range parts {
				got, ok, err := ParseICalDuration(strings.TrimSpace(part))
				require.NoError(t, err)
				require.True(t, ok)
				assert.Equal(t, tt.expected[i], got)
			}
		})
	}
}

func TestParseICalDurationRejectsInvalidRFCForms(t *testing.T) {
	tests := []struct {
		input   string
		wantOK  bool
		wantErr bool
		wantIs  error
	}{
		{input: "", wantOK: false, wantErr: false},
		{input: "1H", wantOK: false, wantErr: false},
		{input: "P", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationMissingDesignator},
		{input: "PT", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationMissingTimeComponent},
		{input: "P1H", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationHoursRequireTimeSection},
		{input: "P1M", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationMinutesRequireTimeSection},
		{input: "P1W2D", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationWeeksOnlyComponent},
		{input: "P1DT2H3M4S5", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationMissingUnit},
		{input: "P1D2H", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationHoursRequireTimeSection},
		{input: "PT1D", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationUnknownUnit},
		{input: "PT1H2D", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationUnknownUnit},
		{input: "PT30M1H", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationDuplicateOrOutOfOrderHoursComponent},
		{input: "PT1H2M3H", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationDuplicateOrOutOfOrderHoursComponent},
		{input: "PT1S30M", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationDuplicateOrOutOfOrderMinutesComponent},
		{input: "PX1D", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationExpectedDigits},
		{input: "P1WT1H", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationWeeksOnlyComponent},
		{input: "PTX1H", wantOK: false, wantErr: true, wantIs: ErrorInvalidICalDurationExpectedDigits},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, ok, err := ParseICalDuration(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantIs != nil {
					assert.True(t, errors.Is(err, tt.wantIs), "expected %v to wrap %v", err, tt.wantIs)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParseDurationAsTimeDuration(t *testing.T) {
	got, ok, err := ParseDurationAsTimeDuration("P15DT5H0M20S")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, 15*24*time.Hour+5*time.Hour+20*time.Second, got)
}

func TestICalDurationAddTo(t *testing.T) {
	start := time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC)
	dur := Duration{Sign: -1, Days: 1, Time: 3 * time.Hour}

	got := dur.AddTo(start)
	want := start.AddDate(0, 0, -1).Add(-3 * time.Hour)
	assert.True(t, want.Equal(got))
}

func TestParseMultiTimeValueShapes(t *testing.T) {
	start := time.Date(2023, 1, 19, 16, 30, 0, 0, time.UTC)
	end := time.Date(2023, 1, 19, 18, 0, 0, 0, time.UTC)

	t.Run("date time", func(t *testing.T) {
		value, err := parseMultiTimeValue("20230119T163000Z", nil)
		require.NoError(t, err)

		single, ok := value.(*DateTimeValue)
		require.True(t, ok)
		assert.True(t, single.StartDate().Equal(start))
		assert.False(t, single.AllDay())
		assert.False(t, single.HasEndDate())
		assert.False(t, single.HasDuration())
		_, ok = single.EndDate()
		assert.False(t, ok)
		_, ok = single.Duration()
		assert.False(t, ok)
	})

	t.Run("date time period", func(t *testing.T) {
		value, err := parseMultiTimeValue("20230119T163000Z/20230119T180000Z", nil)
		require.NoError(t, err)

		period, ok := value.(*DateTimePeriod)
		require.True(t, ok)
		assert.True(t, period.StartDate().Equal(start))
		assert.False(t, period.AllDay())
		gotEnd, ok := period.EndDate()
		require.True(t, ok)
		assert.True(t, gotEnd.Equal(end))
		gotDur, ok := period.Duration()
		require.True(t, ok)
		assert.Equal(t, time.Hour+30*time.Minute, gotDur)
		assert.True(t, period.HasEndDate())
		assert.True(t, period.HasDuration())
	})

	t.Run("date duration", func(t *testing.T) {
		value, err := parseMultiTimeValue("20230119T163000Z/PT1H30M", nil)
		require.NoError(t, err)

		period, ok := value.(*DateDuration)
		require.True(t, ok)
		assert.True(t, period.StartDate().Equal(start))
		assert.False(t, period.AllDay())
		gotEnd, ok := period.EndDate()
		require.True(t, ok)
		assert.True(t, gotEnd.Equal(end))
		gotDur, ok := period.Duration()
		require.True(t, ok)
		assert.Equal(t, time.Hour+30*time.Minute, gotDur)
		assert.True(t, period.HasEndDate())
		assert.True(t, period.HasDuration())
	})

	t.Run("all day date", func(t *testing.T) {
		value, err := parseMultiTimeValue("20230119", map[string][]string{"VALUE": []string{"DATE"}}, ParseAllDay(true))
		require.NoError(t, err)

		single, ok := value.(*DateTimeValue)
		require.True(t, ok)
		got := single.StartDate()
		assert.True(t, single.AllDay())
		assert.Equal(t, 2023, got.Year())
		assert.Equal(t, time.January, got.Month())
		assert.Equal(t, 19, got.Day())
		assert.Equal(t, 0, got.Hour())
		assert.Equal(t, 0, got.Minute())
		assert.Equal(t, 0, got.Second())
	})
}

func TestParseMultiTimeValueRejectsInvalidPeriods(t *testing.T) {
	tests := []string{
		"/20230119T180000Z",
		"20230119T163000Z/",
		"20230119T163000Z/BOGUS",
		"BOGUS",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseMultiTimeValue(input, nil)
			require.Error(t, err)
		})
	}
}

func TestParseTimeValue(t *testing.T) {
	t.Run("timestamp utc", func(t *testing.T) {
		got, err := parseTimeValue("20230119T163000Z", nil, false)
		require.NoError(t, err)
		assert.True(t, got.Equal(time.Date(2023, 1, 19, 16, 30, 0, 0, time.UTC)))
	})

	t.Run("all day date", func(t *testing.T) {
		got, err := parseTimeValue("20230119", map[string][]string{"TZID": []string{"UTC"}}, true)
		require.NoError(t, err)
		assert.Equal(t, 2023, got.Year())
		assert.Equal(t, time.January, got.Month())
		assert.Equal(t, 19, got.Day())
		assert.Equal(t, 0, got.Hour())
	})

	t.Run("timestamp with tzid", func(t *testing.T) {
		got, err := parseTimeValue("20230119T163000", map[string][]string{"TZID": []string{"UTC"}}, false)
		require.NoError(t, err)
		assert.True(t, got.Equal(time.Date(2023, 1, 19, 16, 30, 0, 0, time.UTC)))
	})
}

func TestGetRDateValues(t *testing.T) {
	event := NewEvent("test-rdate-values")
	event.AddRdate("20240101T000000Z,20240102T030000Z/PT2H")

	values, err := event.GetRDateValues("reserved")
	require.NoError(t, err)
	require.Len(t, values, 2)

	first, ok := values[0].(*DateTimeValue)
	require.True(t, ok)
	assert.True(t, first.StartDate().Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.False(t, first.AllDay())

	second, ok := values[1].(*DateDuration)
	require.True(t, ok)
	assert.True(t, second.StartDate().Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)))
	assert.False(t, second.AllDay())
	end, ok := second.EndDate()
	require.True(t, ok)
	assert.True(t, end.Equal(time.Date(2024, 1, 2, 5, 0, 0, 0, time.UTC)))

	dates, err := event.GetRDates()
	require.NoError(t, err)
	require.Len(t, dates, 2)
	assert.True(t, dates[0].Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.True(t, dates[1].Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)))
}

func TestGetRDateValuesStartOnly(t *testing.T) {
	event := NewEvent("test-rdate-values-start-only")
	event.AddRdate("20240102T030000Z/BOGUS")

	values, err := event.GetRDateValues(ParseStartOnly(true))
	require.NoError(t, err)
	require.Len(t, values, 1)

	single, ok := values[0].(*DateTimeValue)
	require.True(t, ok)
	assert.True(t, single.StartDate().Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)))
	assert.False(t, single.AllDay())
}

func TestGetRDateValuesStartOnlyAllDay(t *testing.T) {
	event := NewEvent("test-rdate-values-start-only-allday")
	event.AddRdate("20240102/BOGUS", WithValue(string(ValueDataTypeDate)))

	values, err := event.GetRDateValues(ParseStartOnly(true))
	require.NoError(t, err)
	require.Len(t, values, 1)

	single, ok := values[0].(*DateTimeValue)
	require.True(t, ok)
	assert.True(t, single.AllDay())
	assert.Equal(t, 2024, single.StartDate().Year())
	assert.Equal(t, time.January, single.StartDate().Month())
	assert.Equal(t, 2, single.StartDate().Day())
}

func TestGetRDatesSkipsMalformedPeriodEnd(t *testing.T) {
	event := NewEvent("test-rdate-start-only")
	event.AddRdate("20240102T030000Z/BOGUS")

	dates, err := event.GetRDates()
	require.NoError(t, err)
	require.Len(t, dates, 1)
	assert.True(t, dates[0].Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)))

	_, err = event.GetRDateValues()
	require.Error(t, err)
}

func TestGetExDateValues(t *testing.T) {
	event := NewEvent("test-exdate-values")
	event.AddExdate("20240101T000000Z,20240102T030000Z/20240102T050000Z")

	values, err := event.GetExDateValues()
	require.NoError(t, err)
	require.Len(t, values, 2)

	period, ok := values[1].(*DateTimePeriod)
	require.True(t, ok)
	assert.True(t, period.StartDate().Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)))
	assert.False(t, period.AllDay())
	end, ok := period.EndDate()
	require.True(t, ok)
	assert.True(t, end.Equal(time.Date(2024, 1, 2, 5, 0, 0, 0, time.UTC)))

	dates, err := event.GetExDates()
	require.NoError(t, err)
	require.Len(t, dates, 2)
	assert.True(t, dates[0].Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.True(t, dates[1].Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)))
}

func TestGetExDatesSkipsMalformedPeriodDuration(t *testing.T) {
	event := NewEvent("test-exdate-start-only")
	event.AddExdate("20240102T030000Z/PTBOGUS")

	dates, err := event.GetExDates()
	require.NoError(t, err)
	require.Len(t, dates, 1)
	assert.True(t, dates[0].Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)))

	_, err = event.GetExDateValues()
	require.Error(t, err)
}
