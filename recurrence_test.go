package ics

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRecurrenceRule(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *RecurrenceRule
		wantErr bool
	}{
		{
			name:  "simple daily with count",
			input: "FREQ=DAILY;COUNT=10",
			want: &RecurrenceRule{
				Freq:     FrequencyDaily,
				Count:    10,
				Interval: 1,
			},
		},
		{
			name:  "yearly with BYMONTH and BYDAY",
			input: "FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU",
			want: &RecurrenceRule{
				Freq:     FrequencyYearly,
				Interval: 1,
				ByMonth:  []int{10},
				ByDay:    []WeekdayNum{{OrdWeek: -1, Day: WeekdaySunday}},
			},
		},
		{
			name:  "weekly with interval and multiple days",
			input: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR",
			want: &RecurrenceRule{
				Freq:     FrequencyWeekly,
				Interval: 2,
				ByDay: []WeekdayNum{
					{Day: WeekdayMonday},
					{Day: WeekdayWednesday},
					{Day: WeekdayFriday},
				},
			},
		},
		{
			name:  "monthly with BYMONTHDAY",
			input: "FREQ=MONTHLY;BYMONTHDAY=15",
			want: &RecurrenceRule{
				Freq:       FrequencyMonthly,
				Interval:   1,
				ByMonthDay: []int{15},
			},
		},
		{
			name:  "monthly with ordinal BYDAY",
			input: "FREQ=MONTHLY;BYDAY=2MO",
			want: &RecurrenceRule{
				Freq:     FrequencyMonthly,
				Interval: 1,
				ByDay:    []WeekdayNum{{OrdWeek: 2, Day: WeekdayMonday}},
			},
		},
		{
			name:  "yearly with UNTIL (datetime)",
			input: "FREQ=YEARLY;UNTIL=20251231T235959Z",
			want: &RecurrenceRule{
				Freq:     FrequencyYearly,
				Interval: 1,
				Until:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
		},
		{
			name:  "yearly with UNTIL (date only)",
			input: "FREQ=YEARLY;UNTIL=20251231",
			want: &RecurrenceRule{
				Freq:          FrequencyYearly,
				Interval:      1,
				Until:         time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
				UntilDateOnly: true,
			},
		},
		{
			name:  "with WKST",
			input: "FREQ=WEEKLY;WKST=SU",
			want: &RecurrenceRule{
				Freq:     FrequencyWeekly,
				Interval: 1,
				Wkst:     WeekdaySunday,
			},
		},
		{
			name:  "complex with multiple BY rules",
			input: "FREQ=YEARLY;BYMONTH=3;BYDAY=SU;BYHOUR=2;BYMINUTE=0;BYSECOND=0",
			want: &RecurrenceRule{
				Freq:     FrequencyYearly,
				Interval: 1,
				ByMonth:  []int{3},
				ByDay:    []WeekdayNum{{Day: WeekdaySunday}},
				ByHour:   []int{2},
				ByMinute: []int{0},
				BySecond: []int{0},
			},
		},
		{
			name:  "with BYSETPOS",
			input: "FREQ=MONTHLY;BYDAY=MO,TU,WE,TH,FR;BYSETPOS=-1",
			want: &RecurrenceRule{
				Freq:     FrequencyMonthly,
				Interval: 1,
				ByDay: []WeekdayNum{
					{Day: WeekdayMonday}, {Day: WeekdayTuesday}, {Day: WeekdayWednesday},
					{Day: WeekdayThursday}, {Day: WeekdayFriday},
				},
				BySetPos: []int{-1},
			},
		},
		{
			name:  "with BYYEARDAY and BYWEEKNO",
			input: "FREQ=YEARLY;BYYEARDAY=1,100,200;BYWEEKNO=20",
			want: &RecurrenceRule{
				Freq:      FrequencyYearly,
				Interval:  1,
				ByYearDay: []int{1, 100, 200},
				ByWeekNo:  []int{20},
			},
		},
		{
			name:    "missing FREQ",
			input:   "COUNT=10",
			wantErr: true,
		},
		{
			name:    "invalid FREQ",
			input:   "FREQ=BIWEEKLY",
			wantErr: true,
		},
		{
			name:    "invalid BYDAY",
			input:   "FREQ=WEEKLY;BYDAY=XX",
			wantErr: true,
		},
		{
			name:    "invalid COUNT",
			input:   "FREQ=DAILY;COUNT=abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:  "unknown keys are ignored per RFC 5545",
			input: "FREQ=DAILY;X-CUSTOM=foo;COUNT=5",
			want: &RecurrenceRule{
				Freq:     FrequencyDaily,
				Count:    5,
				Interval: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRecurrenceRule(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Freq, got.Freq)
			assert.Equal(t, tt.want.Count, got.Count)
			assert.Equal(t, tt.want.Interval, got.Interval)
			assert.Equal(t, tt.want.ByMonth, got.ByMonth)
			assert.Equal(t, tt.want.ByDay, got.ByDay)
			assert.Equal(t, tt.want.ByMonthDay, got.ByMonthDay)
			assert.Equal(t, tt.want.ByYearDay, got.ByYearDay)
			assert.Equal(t, tt.want.ByWeekNo, got.ByWeekNo)
			assert.Equal(t, tt.want.ByHour, got.ByHour)
			assert.Equal(t, tt.want.ByMinute, got.ByMinute)
			assert.Equal(t, tt.want.BySecond, got.BySecond)
			assert.Equal(t, tt.want.BySetPos, got.BySetPos)
			assert.Equal(t, tt.want.Wkst, got.Wkst)
			if !tt.want.Until.IsZero() {
				assert.True(t, tt.want.Until.Equal(got.Until), "Until: want %v, got %v", tt.want.Until, got.Until)
				assert.Equal(t, tt.want.UntilDateOnly, got.UntilDateOnly)
			}
		})
	}
}

func TestRecurrenceRuleString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple daily",
			input: "FREQ=DAILY;COUNT=10",
			want:  "FREQ=DAILY;COUNT=10",
		},
		{
			name:  "yearly with BYMONTH and BYDAY",
			input: "FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU",
			want:  "FREQ=YEARLY;BYDAY=-1SU;BYMONTH=10",
		},
		{
			name:  "weekly with interval",
			input: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR",
			want:  "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR",
		},
		{
			name:  "with UNTIL datetime",
			input: "FREQ=YEARLY;UNTIL=20251231T235959Z",
			want:  "FREQ=YEARLY;UNTIL=20251231T235959Z",
		},
		{
			name:  "with UNTIL date only",
			input: "FREQ=YEARLY;UNTIL=20251231",
			want:  "FREQ=YEARLY;UNTIL=20251231",
		},
		{
			name:  "with WKST",
			input: "FREQ=WEEKLY;WKST=SU",
			want:  "FREQ=WEEKLY;WKST=SU",
		},
		{
			name:  "midnight UTC UNTIL round-trips as datetime not date-only",
			input: "FREQ=YEARLY;UNTIL=20251231T000000Z",
			want:  "FREQ=YEARLY;UNTIL=20251231T000000Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseRecurrenceRule(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, rule.String())
		})
	}
}

func TestGetRRules(t *testing.T) {
	event := NewEvent("test-rrule")
	event.AddRrule("FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU")

	rules, err := event.GetRRules()
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, FrequencyYearly, rules[0].Freq)
	assert.Equal(t, []int{10}, rules[0].ByMonth)
	assert.Equal(t, []WeekdayNum{{OrdWeek: -1, Day: WeekdaySunday}}, rules[0].ByDay)
}

func TestGetRRulesMultiple(t *testing.T) {
	event := NewEvent("test-rrule-multi")
	event.AddRrule("FREQ=DAILY;COUNT=5")
	event.AddRrule("FREQ=WEEKLY;BYDAY=MO")

	rules, err := event.GetRRules()
	require.NoError(t, err)
	require.Len(t, rules, 2)
	assert.Equal(t, FrequencyDaily, rules[0].Freq)
	assert.Equal(t, FrequencyWeekly, rules[1].Freq)
}

func TestGetRRulesEmpty(t *testing.T) {
	event := NewEvent("test-no-rrule")
	rules, err := event.GetRRules()
	require.NoError(t, err)
	assert.Nil(t, rules)
}

func TestGetExRules(t *testing.T) {
	event := NewEvent("test-exrule")
	event.AddExrule("FREQ=YEARLY;BYMONTH=12;BYMONTHDAY=25")

	rules, err := event.GetExRules()
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, FrequencyYearly, rules[0].Freq)
}

func TestGetExDates(t *testing.T) {
	event := NewEvent("test-exdate")
	event.AddExdate("20231225T120000Z")

	dates, err := event.GetExDates()
	require.NoError(t, err)
	require.Len(t, dates, 1)
	assert.True(t, dates[0].Equal(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)))
}

func TestGetExDatesCommaSeparated(t *testing.T) {
	event := NewEvent("test-exdate-csv")
	event.AddExdate("20231225T120000Z,20231226T120000Z")

	dates, err := event.GetExDates()
	require.NoError(t, err)
	require.Len(t, dates, 2)
	assert.True(t, dates[0].Equal(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)))
	assert.True(t, dates[1].Equal(time.Date(2023, 12, 26, 12, 0, 0, 0, time.UTC)))
}

func TestGetExDatesMultipleProperties(t *testing.T) {
	event := NewEvent("test-exdate-multi")
	event.AddExdate("20231225T120000Z")
	event.AddExdate("20231226T120000Z")

	dates, err := event.GetExDates()
	require.NoError(t, err)
	require.Len(t, dates, 2)
}

func TestGetExDatesAllDay(t *testing.T) {
	event := NewEvent("test-exdate-allday")
	event.AddExdate("20231225")

	dates, err := event.GetExDates()
	require.NoError(t, err)
	require.Len(t, dates, 1)
	assert.Equal(t, 2023, dates[0].Year())
	assert.Equal(t, time.December, dates[0].Month())
	assert.Equal(t, 25, dates[0].Day())
}

func TestGetExDatesEmpty(t *testing.T) {
	event := NewEvent("test-no-exdate")
	dates, err := event.GetExDates()
	require.NoError(t, err)
	assert.Nil(t, dates)
}

func TestGetRDates(t *testing.T) {
	event := NewEvent("test-rdate")
	event.AddRdate("20240101T000000Z")

	dates, err := event.GetRDates()
	require.NoError(t, err)
	require.Len(t, dates, 1)
	assert.True(t, dates[0].Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
}

func TestGetRDatesCommaSeparated(t *testing.T) {
	event := NewEvent("test-rdate-csv")
	event.AddRdate("20240101T000000Z,20240201T000000Z")

	dates, err := event.GetRDates()
	require.NoError(t, err)
	require.Len(t, dates, 2)
}

func TestGetRecurrenceID(t *testing.T) {
	ical := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:test-recurrence-id
RECURRENCE-ID:20231225T120000Z
DTSTART:20231226T120000Z
DTEND:20231226T130000Z
SUMMARY:Modified occurrence
END:VEVENT
END:VCALENDAR`

	cal, err := ParseCalendar(strings.NewReader(ical))
	require.NoError(t, err)

	events := cal.Events()
	require.Len(t, events, 1)

	recID, err := events[0].GetRecurrenceID()
	require.NoError(t, err)
	assert.True(t, recID.Equal(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)))
}

func TestGetRecurrenceIDNotSet(t *testing.T) {
	event := NewEvent("test-no-recurrence-id")
	_, err := event.GetRecurrenceID()
	assert.Error(t, err)
}

func TestGetExDatesWithTZID(t *testing.T) {
	ical := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:test-exdate-tzid
DTSTART;TZID=America/New_York:20231201T090000
EXDATE;TZID=America/New_York:20231208T090000
SUMMARY:Weekly meeting
RRULE:FREQ=WEEKLY;COUNT=10
END:VEVENT
END:VCALENDAR`

	cal, err := ParseCalendar(strings.NewReader(ical))
	require.NoError(t, err)

	events := cal.Events()
	require.Len(t, events, 1)

	dates, err := events[0].GetExDates()
	require.NoError(t, err)
	require.Len(t, dates, 1)

	loc, _ := time.LoadLocation("America/New_York")
	expected := time.Date(2023, 12, 8, 9, 0, 0, 0, loc)
	assert.True(t, dates[0].Equal(expected), "want %v, got %v", expected, dates[0])
}
