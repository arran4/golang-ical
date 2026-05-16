package ics

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Frequency string

const (
	FrequencySecondly Frequency = "SECONDLY"
	FrequencyMinutely Frequency = "MINUTELY"
	FrequencyHourly   Frequency = "HOURLY"
	FrequencyDaily    Frequency = "DAILY"
	FrequencyWeekly   Frequency = "WEEKLY"
	FrequencyMonthly  Frequency = "MONTHLY"
	FrequencyYearly   Frequency = "YEARLY"
)

type Weekday string

const (
	WeekdaySunday    Weekday = "SU"
	WeekdayMonday    Weekday = "MO"
	WeekdayTuesday   Weekday = "TU"
	WeekdayWednesday Weekday = "WE"
	WeekdayThursday  Weekday = "TH"
	WeekdayFriday    Weekday = "FR"
	WeekdaySaturday  Weekday = "SA"
)

type rruleKey string

const (
	rruleKeyFreq       rruleKey = "FREQ"
	rruleKeyUntil      rruleKey = "UNTIL"
	rruleKeyCount      rruleKey = "COUNT"
	rruleKeyInterval   rruleKey = "INTERVAL"
	rruleKeyBySecond   rruleKey = "BYSECOND"
	rruleKeyByMinute   rruleKey = "BYMINUTE"
	rruleKeyByHour     rruleKey = "BYHOUR"
	rruleKeyByDay      rruleKey = "BYDAY"
	rruleKeyByMonthDay rruleKey = "BYMONTHDAY"
	rruleKeyByYearDay  rruleKey = "BYYEARDAY"
	rruleKeyByWeekNo   rruleKey = "BYWEEKNO"
	rruleKeyByMonth    rruleKey = "BYMONTH"
	rruleKeyBySetPos   rruleKey = "BYSETPOS"
	rruleKeyWkst       rruleKey = "WKST"
)

var validWeekdays = map[Weekday]bool{
	WeekdaySunday: true, WeekdayMonday: true, WeekdayTuesday: true,
	WeekdayWednesday: true, WeekdayThursday: true, WeekdayFriday: true,
	WeekdaySaturday: true,
}

// WeekdayNum represents a weekday with an optional ordinal (e.g., -1SU = last Sunday, 2MO = second Monday).
// OrdWeek of 0 means no ordinal was specified.
type WeekdayNum struct {
	OrdWeek int
	Day     Weekday
}

func (wdn WeekdayNum) String() string {
	if wdn.OrdWeek == 0 {
		return string(wdn.Day)
	}
	return fmt.Sprintf("%d%s", wdn.OrdWeek, wdn.Day)
}

// RecurrenceRule represents a parsed RRULE as defined in RFC 5545 Section 3.3.10.
type RecurrenceRule struct {
	Freq          Frequency
	Until         time.Time // zero value if unset
	UntilDateOnly bool      // true when UNTIL was parsed from a date-only value (no time component)
	Count         int       // 0 if unset
	Interval      int       // defaults to 1
	BySecond      []int
	ByMinute      []int
	ByHour        []int
	ByDay         []WeekdayNum
	ByMonthDay    []int
	ByYearDay     []int
	ByWeekNo      []int
	ByMonth       []int
	BySetPos      []int
	Wkst          Weekday // empty string if unset
}

// ParseRecurrenceRule parses an RRULE value string (without the "RRULE:" prefix)
// into a RecurrenceRule struct.
func ParseRecurrenceRule(s string) (*RecurrenceRule, error) {
	rule := &RecurrenceRule{
		Interval: 1,
	}

	parts := strings.Split(s, ";")
	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid RRULE part: %q", part)
		}
		key, value := kv[0], kv[1]

		var err error
		switch rruleKey(key) {
		case rruleKeyFreq:
			rule.Freq = Frequency(value)
			if !isValidFrequency(rule.Freq) {
				return nil, fmt.Errorf("invalid FREQ value: %q", value)
			}
		case rruleKeyUntil:
			rule.Until, rule.UntilDateOnly, err = parseRecurrenceTime(value)
			if err != nil {
				return nil, fmt.Errorf("invalid UNTIL value: %w", err)
			}
		case rruleKeyCount:
			rule.Count, err = strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid COUNT value: %w", err)
			}
		case rruleKeyInterval:
			rule.Interval, err = strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid INTERVAL value: %w", err)
			}
		case rruleKeyBySecond:
			rule.BySecond, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYSECOND value: %w", err)
			}
		case rruleKeyByMinute:
			rule.ByMinute, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYMINUTE value: %w", err)
			}
		case rruleKeyByHour:
			rule.ByHour, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYHOUR value: %w", err)
			}
		case rruleKeyByDay:
			rule.ByDay, err = parseWeekdayNumList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYDAY value: %w", err)
			}
		case rruleKeyByMonthDay:
			rule.ByMonthDay, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYMONTHDAY value: %w", err)
			}
		case rruleKeyByYearDay:
			rule.ByYearDay, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYYEARDAY value: %w", err)
			}
		case rruleKeyByWeekNo:
			rule.ByWeekNo, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYWEEKNO value: %w", err)
			}
		case rruleKeyByMonth:
			rule.ByMonth, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYMONTH value: %w", err)
			}
		case rruleKeyBySetPos:
			rule.BySetPos, err = parseIntList(value)
			if err != nil {
				return nil, fmt.Errorf("invalid BYSETPOS value: %w", err)
			}
		case rruleKeyWkst:
			rule.Wkst = Weekday(value)
			if !validWeekdays[rule.Wkst] {
				return nil, fmt.Errorf("invalid WKST value: %q", value)
			}
		default:
			// RFC 5545 says implementations SHOULD ignore unrecognized properties.
			// This handles vendor extensions (X-...) and future RFC additions.
		}
	}

	if rule.Freq == "" {
		return nil, fmt.Errorf("RRULE missing required FREQ")
	}

	return rule, nil
}

// String serializes the RecurrenceRule back to RRULE value format (without "RRULE:" prefix).
func (r *RecurrenceRule) String() string {
	var parts []string

	parts = append(parts, "FREQ="+string(r.Freq))

	if !r.Until.IsZero() {
		if r.UntilDateOnly {
			parts = append(parts, "UNTIL="+r.Until.Format(icalDateFormatLocal))
		} else {
			parts = append(parts, "UNTIL="+r.Until.Format(icalTimestampFormatUtc))
		}
	}
	if r.Count != 0 {
		parts = append(parts, "COUNT="+strconv.Itoa(r.Count))
	}
	if r.Interval != 0 && r.Interval != 1 {
		parts = append(parts, "INTERVAL="+strconv.Itoa(r.Interval))
	}
	if len(r.BySecond) > 0 {
		parts = append(parts, "BYSECOND="+intListString(r.BySecond))
	}
	if len(r.ByMinute) > 0 {
		parts = append(parts, "BYMINUTE="+intListString(r.ByMinute))
	}
	if len(r.ByHour) > 0 {
		parts = append(parts, "BYHOUR="+intListString(r.ByHour))
	}
	if len(r.ByDay) > 0 {
		strs := make([]string, len(r.ByDay))
		for i, wd := range r.ByDay {
			strs[i] = wd.String()
		}
		parts = append(parts, "BYDAY="+strings.Join(strs, ","))
	}
	if len(r.ByMonthDay) > 0 {
		parts = append(parts, "BYMONTHDAY="+intListString(r.ByMonthDay))
	}
	if len(r.ByYearDay) > 0 {
		parts = append(parts, "BYYEARDAY="+intListString(r.ByYearDay))
	}
	if len(r.ByWeekNo) > 0 {
		parts = append(parts, "BYWEEKNO="+intListString(r.ByWeekNo))
	}
	if len(r.ByMonth) > 0 {
		parts = append(parts, "BYMONTH="+intListString(r.ByMonth))
	}
	if len(r.BySetPos) > 0 {
		parts = append(parts, "BYSETPOS="+intListString(r.BySetPos))
	}
	if r.Wkst != "" {
		parts = append(parts, "WKST="+string(r.Wkst))
	}

	return strings.Join(parts, ";")
}

func isValidFrequency(f Frequency) bool {
	switch f {
	case FrequencySecondly, FrequencyMinutely, FrequencyHourly,
		FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly:
		return true
	}
	return false
}

func parseRecurrenceTime(s string) (time.Time, bool, error) {
	// Try datetime with UTC (20060102T150405Z)
	if t, err := time.Parse(icalTimestampFormatUtc, s); err == nil {
		return t, false, nil
	}
	// Try date only with UTC (20060102Z)
	if t, err := time.Parse(icalDateFormatUtc, s); err == nil {
		return t, true, nil
	}
	// Try date only without timezone (20060102)
	if t, err := time.Parse(icalDateFormatLocal, s); err == nil {
		return t, true, nil
	}
	// Try datetime without UTC (20060102T150405)
	if t, err := time.Parse(icalTimestampFormatLocal, s); err == nil {
		return t, false, nil
	}
	return time.Time{}, false, fmt.Errorf("cannot parse time %q", s)
}

func parseIntList(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, nil
}

func parseWeekdayNumList(s string) ([]WeekdayNum, error) {
	parts := strings.Split(s, ",")
	result := make([]WeekdayNum, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		wdn, err := parseWeekdayNum(p)
		if err != nil {
			return nil, err
		}
		result = append(result, wdn)
	}
	return result, nil
}

func parseWeekdayNum(s string) (WeekdayNum, error) {
	if len(s) < 2 {
		return WeekdayNum{}, fmt.Errorf("invalid BYDAY value: %q", s)
	}

	day := Weekday(s[len(s)-2:])
	if !validWeekdays[day] {
		return WeekdayNum{}, fmt.Errorf("invalid weekday: %q", s)
	}

	ordStr := s[:len(s)-2]
	if ordStr == "" {
		return WeekdayNum{Day: day}, nil
	}

	ord, err := strconv.Atoi(ordStr)
	if err != nil {
		return WeekdayNum{}, fmt.Errorf("invalid ordinal in BYDAY value %q: %w", s, err)
	}

	return WeekdayNum{OrdWeek: ord, Day: day}, nil
}

func intListString(nums []int) string {
	strs := make([]string, len(nums))
	for i, n := range nums {
		strs[i] = strconv.Itoa(n)
	}
	return strings.Join(strs, ",")
}
