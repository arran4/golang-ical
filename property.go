package ics

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type BaseProperty struct {
	IANAToken      string
	ICalParameters map[string][]string
	Value          string
}

type PropertyParameter interface {
	KeyValue(s ...interface{}) (string, []string)
}

type KeyValues struct {
	Key   string
	Value []string
}

func (kv *KeyValues) KeyValue(_ ...interface{}) (string, []string) {
	return kv.Key, kv.Value
}

func WithCN(cn string) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterCn),
		Value: []string{cn},
	}
}

func WithTZID(tzid string) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterTzid),
		Value: []string{tzid},
	}
}

// WithAlternativeRepresentation takes what must be a valid URI in quotation marks
func WithAlternativeRepresentation(uri *url.URL) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterAltrep),
		Value: []string{uri.String()},
	}
}

func WithEncoding(encType string) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterEncoding),
		Value: []string{encType},
	}
}

func WithFmtType(contentType string) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterFmttype),
		Value: []string{contentType},
	}
}

func WithValue(kind string) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterValue),
		Value: []string{kind},
	}
}

func WithRSVP(b bool) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterRsvp),
		Value: []string{strconv.FormatBool(b)},
	}
}

func trimUT8StringUpTo(maxLength int, s string) string {
	length := 0
	lastWordBoundary := -1
	var lastRune rune
	for i, r := range s {
		if r == ' ' || r == '<' {
			lastWordBoundary = i
		} else if lastRune == '>' {
			lastWordBoundary = i
		}
		lastRune = r
		newLength := length + utf8.RuneLen(r)
		if newLength > maxLength {
			break
		}
		length = newLength
	}
	if lastWordBoundary > 0 {
		return s[:lastWordBoundary]
	}

	return s[:length]
}

func (bp *BaseProperty) parameterValue(param Parameter) (string, error) {
	v, ok := bp.ICalParameters[string(param)]
	if !ok || len(v) == 0 {
		return "", fmt.Errorf("parameter %q not found in property", param)
	}
	if len(v) != 1 {
		return "", fmt.Errorf("expected only one value for parameter %q in property, found %d", param, len(v))
	}
	return v[0], nil
}

func (bp *BaseProperty) GetValueType() ValueDataType {
	for k, v := range bp.ICalParameters {
		if Parameter(k) == ParameterValue && len(v) == 1 {
			return ValueDataType(v[0])
		}
	}

	// defaults from spec if unspecified
	switch Property(bp.IANAToken) {
	default:
		fallthrough
	case PropertyCalscale, PropertyMethod, PropertyProductId, PropertyVersion, PropertyCategories, PropertyClass,
		PropertyComment, PropertyDescription, PropertyLocation, PropertyResources, PropertyStatus, PropertySummary,
		PropertyTransp, PropertyTzid, PropertyTzname, PropertyContact, PropertyRelatedTo, PropertyUid, PropertyAction,
		PropertyRequestStatus:
		return ValueDataTypeText

	case PropertyAttach, PropertyTzurl, PropertyUrl:
		return ValueDataTypeUri

	case PropertyGeo:
		return ValueDataTypeFloat

	case PropertyPercentComplete, PropertyPriority, PropertyRepeat, PropertySequence:
		return ValueDataTypeInteger

	case PropertyCompleted, PropertyDtend, PropertyDue, PropertyDtstart, PropertyRecurrenceId, PropertyExdate,
		PropertyRdate, PropertyCreated, PropertyDtstamp, PropertyLastModified:
		return ValueDataTypeDateTime

	case PropertyDuration, PropertyTrigger:
		return ValueDataTypeDuration

	case PropertyFreebusy:
		return ValueDataTypePeriod

	case PropertyTzoffsetfrom, PropertyTzoffsetto:
		return ValueDataTypeUtcOffset

	case PropertyAttendee, PropertyOrganizer:
		return ValueDataTypeCalAddress

	case PropertyRrule:
		return ValueDataTypeRecur
	}
}

func (bp *BaseProperty) serialize(w io.Writer, serialConfig *SerializationConfiguration) error {
	b := bytes.NewBufferString("")
	_, _ = fmt.Fprint(b, bp.IANAToken)

	var keys []string
	for k := range bp.ICalParameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := bp.ICalParameters[k]
		_, _ = fmt.Fprint(b, ";")
		_, _ = fmt.Fprint(b, k)
		_, _ = fmt.Fprint(b, "=")
		for vi, v := range vs {
			if vi > 0 {
				_, _ = fmt.Fprint(b, ",")
			}
			if Parameter(k).IsQuoted() {
				v = quotedValueString(v)
				_, _ = fmt.Fprint(b, v)
			} else {
				v = escapeValueString(v)
				_, _ = fmt.Fprint(b, v)
			}
		}
	}
	_, _ = fmt.Fprint(b, ":")
	propertyValue := bp.Value
	if bp.GetValueType() == ValueDataTypeText {
		propertyValue = ToText(propertyValue)
	}
	_, _ = fmt.Fprint(b, propertyValue)
	r := b.String()
	if len(r) > serialConfig.MaxLength {
		l := trimUT8StringUpTo(serialConfig.MaxLength, r)
		_, err := fmt.Fprint(w, l, serialConfig.NewLine)
		if err != nil {
			return fmt.Errorf("property %s serialization: %w", bp.IANAToken, err)
		}
		r = r[len(l):]

		for len(r) > serialConfig.MaxLength-1 {
			l := trimUT8StringUpTo(serialConfig.MaxLength-1, r)
			_, err = fmt.Fprint(w, " ", l, serialConfig.NewLine)
			if err != nil {
				return fmt.Errorf("property %s serialization: %w", bp.IANAToken, err)
			}
			r = r[len(l):]
		}
		_, err = fmt.Fprint(w, " ")
		if err != nil {
			return fmt.Errorf("property %s serialization: %w", bp.IANAToken, err)
		}
	}
	_, err := fmt.Fprint(w, r, serialConfig.NewLine)
	if err != nil {
		return fmt.Errorf("property %s serialization: %w", bp.IANAToken, err)
	}
	return nil
}

func escapeValueString(v string) string {
	changed := 0
	result := ""
	for i, r := range v {
		switch r {
		case ',', '"', ';', ':', '\\', '\'':
			result = result + v[changed:i] + "\\" + string(r)
			changed = i + 1
		}
	}
	if changed == 0 {
		return v
	}
	return result + v[changed:]
}

func quotedValueString(v string) string {
	changed := 0
	result := ""
	for i, r := range v {
		switch r {
		case '"', '\\':
			result = result + v[changed:i] + "\\" + string(r)
			changed = i + 1
		}
	}
	if changed == 0 {
		return `"` + v + `"`
	}
	return `"` + result + v[changed:] + `"`
}

type IANAProperty struct {
	BaseProperty
}

// ParseTime Parses the time, all day is if we should treat the value as an all day event.
// Returns the time if parsable, if it is an all day time, and an error if there is one
func (p IANAProperty) ParseTime(expectAllDay bool) (*time.Time, bool, error) {
	timeVal := p.BaseProperty.Value
	matched := timeStampVariations.FindStringSubmatch(timeVal)
	if matched == nil {
		return nil, false, fmt.Errorf("time value not matched, got '%s'", timeVal)
	}
	tOrZGrp := matched[2]
	zGrp := matched[4]
	grp1len := len(matched[1])
	grp3len := len(matched[3])

	tzId, tzIdOk := p.ICalParameters["TZID"]
	var propLoc *time.Location
	if tzIdOk {
		if len(tzId) != 1 {
			return nil, false, errors.New("expected only one TZID")
		}
		var tzErr error
		propLoc, tzErr = time.LoadLocation(tzId[0])
		if tzErr != nil {
			return nil, false, tzErr
		}
	}
	dateStr := matched[1]

	if expectAllDay {
		if grp1len > 0 {
			if tOrZGrp == "Z" || zGrp == "Z" {
				t, err := time.ParseInLocation(icalDateFormatUtc, dateStr+"Z", time.UTC)
				return &t, true, err
			} else {
				if propLoc == nil {
					t, err := time.ParseInLocation(icalDateFormatLocal, dateStr, time.Local)
					return &t, true, err
				} else {
					t, err := time.ParseInLocation(icalDateFormatLocal, dateStr, propLoc)
					return &t, true, err
				}
			}
		}
		return nil, false, fmt.Errorf("time value matched but unsupported all-day timestamp, got '%s'", timeVal)
	}

	switch {
	case grp1len > 0 && grp3len > 0 && tOrZGrp == "T" && zGrp == "Z":
		t, err := time.ParseInLocation(icalTimestampFormatUtc, timeVal, time.UTC)
		return &t, false, err
	case grp1len > 0 && grp3len > 0 && tOrZGrp == "T" && zGrp == "":
		if propLoc == nil {
			t, err := time.ParseInLocation(icalTimestampFormatLocal, timeVal, time.Local)
			return &t, false, err
		} else {
			t, err := time.ParseInLocation(icalTimestampFormatLocal, timeVal, propLoc)
			return &t, false, err
		}
	case grp1len > 0 && grp3len == 0 && tOrZGrp == "Z" && zGrp == "":
		t, err := time.ParseInLocation(icalDateFormatUtc, dateStr+"Z", time.UTC)
		return &t, true, err
	case grp1len > 0 && grp3len == 0 && tOrZGrp == "" && zGrp == "":
		if propLoc == nil {
			t, err := time.ParseInLocation(icalDateFormatLocal, dateStr, time.Local)
			return &t, true, err
		} else {
			t, err := time.ParseInLocation(icalDateFormatLocal, dateStr, propLoc)
			return &t, true, err
		}
	}

	return nil, false, fmt.Errorf("time value matched but not supported, got '%s'", timeVal)
}

// ParseDurations assumes the value is a duration and tries to parse it
//
//	 Value Name:  DURATION
//
//	Purpose:  This value type is used to identify properties that contain
//	   a duration of time.
//
//	Format Definition:  This value type is defined by the following
//	   notation:
//
//	    dur-value  = (["+"] / "-") "P" (dur-date / dur-time / dur-week)
//
//	    dur-date   = dur-day [dur-time]
//	    dur-time   = "T" (dur-hour / dur-minute / dur-second)
//	    dur-week   = 1*DIGIT "W"
//	    dur-hour   = 1*DIGIT "H" [dur-minute]
//	    dur-minute = 1*DIGIT "M" [dur-second]
//	    dur-second = 1*DIGIT "S"
//	    dur-day    = 1*DIGIT "D"
//
//	Description:  If the property permits, multiple "duration" values are
//	   specified by a COMMA-separated list of values.  The format is
//	   based on the [ISO.8601.2004] complete representation basic format
//	   with designators for the duration of time.  The format can
//	   represent nominal durations (weeks and days) and accurate
//	   durations (hours, minutes, and seconds).  Note that unlike
//	   [ISO.8601.2004], this value type doesn't support the "Y" and "M"
//	   designators to specify durations in terms of years and months.
//
// Desruisseaux                Standards Track                    [Page 35]
//
// # RFC 5545                       iCalendar                  September 2009
//
//	   The duration of a week or a day depends on its position in the
//	   calendar.  In the case of discontinuities in the time scale, such
//	   as the change from standard time to daylight time and back, the
//	   computation of the exact duration requires the subtraction or
//	   addition of the change of duration of the discontinuity.  Leap
//	   seconds MUST NOT be considered when computing an exact duration.
//	   When computing an exact duration, the greatest order time
//	   components MUST be added first, that is, the number of days MUST
//	   be added first, followed by the number of hours, number of
//	   minutes, and number of seconds.
//
//	   Negative durations are typically used to schedule an alarm to
//	   trigger before an associated time (see Section 3.8.6.3).
//
//	   No additional content value encoding (i.e., BACKSLASH character
//	   encoding, see Section 3.3.11) are defined for this value type.
//
//	Example:  A duration of 15 days, 5 hours, and 20 seconds would be:
//
//	    P15DT5H0M20S
//
//	   A duration of 7 weeks would be:
//
//	    P7W
func (p IANAProperty) ParseDurations() ([]Duration, error) {
	var result []Duration
	br := bytes.NewReader([]byte(strings.ToUpper(p.Value)))
	for {
		value, err := ParseDurationReader(br)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("%w: '%s'", err, p.Value)
		}
		if value != nil {
			result = append(result, *value)
		}
		if err == io.EOF {
			return result, nil
		}
	}
}

type DurationOrder struct {
	Key      rune
	Value    *Duration
	Required bool
}

var order = []DurationOrder{
	{Key: 'P', Value: nil, Required: true},
	{Key: 'W', Value: &Duration{Duration: 0, Days: 7}},
	{Key: 'D', Value: &Duration{Duration: 0, Days: 1}},
	{Key: 'T', Value: nil},
	{Key: 'H', Value: &Duration{Duration: time.Hour, Days: 0}},
	{Key: 'M', Value: &Duration{Duration: time.Minute, Days: 0}},
	{Key: 'S', Value: &Duration{Duration: time.Second, Days: 0}},
}

func ParseDuration(s string) (*Duration, error) {
	return ParseDurationReader(strings.NewReader(strings.ToUpper(s)))
}

type ReaderRuneBuffer interface {
	ReadRune() (rune, int, error)
	UnreadRune() error
}

func ParseDurationReader(br ReaderRuneBuffer) (*Duration, error) {
	var value = Duration{
		Positive: true,
	}
	pos := 0
	for pos != 1 {
		b, _, err := br.ReadRune()
		if err == io.EOF {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse duration")
		}
		switch b {
		case '-':
			value.Positive = false
		case '+':
		case 'P':
			pos = 1
		default:
			return nil, fmt.Errorf("missing p initializer got %c", b)
		}
	}
	for pos < len(order) {
		var number int
		var b rune
		var err error
		for {
			b, _, err = br.ReadRune()
			if err == io.EOF || b == ',' {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to parse duration")
			}
			if unicode.IsSpace(b) {
				continue
			}
			if unicode.IsDigit(b) {
				number = number*10 + int(b-'0')
			} else {
				break
			}
		}
		if err == io.EOF || b == ',' {
			break
		}
		for ; pos < len(order) && order[pos].Key != b; pos++ {
		}
		if pos >= len(order) {
			err := br.UnreadRune()
			if err != nil {
				return nil, fmt.Errorf("unread rune error '%w'", err)
			}
			break
		}
		selected := order[pos]
		if selected.Value != nil {
			value.Days += selected.Value.Days * number
			value.Duration += selected.Value.Duration * time.Duration(number)
		}
	}
	return &value, nil
}

type Duration struct {
	Positive bool
	Duration time.Duration
	Days     int
}

var (
	propertyIanaTokenReg *regexp.Regexp
	propertyParamNameReg *regexp.Regexp
	propertyValueTextReg *regexp.Regexp
)

func init() {
	var err error
	propertyIanaTokenReg, err = regexp.Compile("[A-Za-z0-9-]{1,}")
	if err != nil {
		log.Panicf("Failed to build regex: %v", err)
	}
	propertyParamNameReg = propertyIanaTokenReg
	propertyValueTextReg, err = regexp.Compile("^.*")
	if err != nil {
		log.Panicf("Failed to build regex: %v", err)
	}
}

type ContentLine string

func ParseProperty(contentLine ContentLine) (*BaseProperty, error) {
	r := &BaseProperty{
		ICalParameters: map[string][]string{},
	}
	tokenPos := propertyIanaTokenReg.FindIndex([]byte(contentLine))
	if tokenPos == nil {
		return nil, nil
	}
	p := 0
	r.IANAToken = string(contentLine[p+tokenPos[0] : p+tokenPos[1]])
	p += tokenPos[1]
	for {
		if p >= len(contentLine) {
			return nil, nil
		}
		switch rune(contentLine[p]) {
		case ':':
			return parsePropertyValue(r, string(contentLine), p+1), nil
		case ';':
			var np int
			var err error
			t := r.IANAToken
			r, np, err = parsePropertyParam(r, string(contentLine), p+1)
			if err != nil {
				return nil, fmt.Errorf("parsing property %s: %w", t, err)
			}
			if r == nil {
				return nil, nil
			}
			p = np
		default:
			return nil, nil
		}
	}
}

func parsePropertyParam(r *BaseProperty, contentLine string, p int) (*BaseProperty, int, error) {
	tokenPos := propertyParamNameReg.FindIndex([]byte(contentLine[p:]))
	if tokenPos == nil {
		return nil, p, nil
	}
	k, v := "", ""
	k = string(contentLine[p : p+tokenPos[1]])
	p += tokenPos[1]
	if p >= len(contentLine) {
		return nil, p, fmt.Errorf("missing property param operator for %s in %s", k, r.IANAToken)
	}
	switch rune(contentLine[p]) {
	case '=':
		p += 1
	default:
		return nil, p, fmt.Errorf("missing property value for %s in %s", k, r.IANAToken)
	}
	for {
		if p >= len(contentLine) {
			return nil, p, nil
		}
		var err error
		v, p, err = parsePropertyParamValue(contentLine, p)
		if err != nil {
			return nil, 0, fmt.Errorf("parse error: %w %s in %s", err, k, r.IANAToken)
		}
		r.ICalParameters[k] = append(r.ICalParameters[k], v)
		if p >= len(contentLine) {
			return nil, p, fmt.Errorf("unexpected end of property %s", r.IANAToken)
		}
		switch rune(contentLine[p]) {
		case ',':
			p += 1
		default:
			return r, p, nil
		}
	}
}

func parsePropertyParamValue(s string, p int) (string, int, error) {
	/*
	   quoted-string = DQUOTE *QSAFE-CHAR DQUOTE

	   QSAFE-CHAR    = WSP / %x21 / %x23-7E / NON-US-ASCII
	   ; Any character except CONTROL and DQUOTE

	   SAFE-CHAR     = WSP / %x21 / %x23-2B / %x2D-39 / %x3C-7E
	                 / NON-US-ASCII
	   ; Any character except CONTROL, DQUOTE, ";", ":", ","

	   text       = *(TSAFE-CHAR / ":" / DQUOTE / ESCAPED-CHAR)
	   ; Folded according to description above

	   ESCAPED-CHAR = "\\" / "\;" / "\," / "\N" / "\n")
	      ; \\ encodes \, \N or \n encodes newline
	      ; \; encodes ;, \, encodes ,

	   TSAFE-CHAR = %x20-21 / %x23-2B / %x2D-39 / %x3C-5B
	                %x5D-7E / NON-US-ASCII
	      ; Any character except CTLs not needed by the current
	      ; character set, DQUOTE, ";", ":", "\", ","

	   CONTROL       = %x00-08 / %x0A-1F / %x7F
	   ; All the controls except HTAB

	*/
	r := make([]byte, 0, len(s))
	quoted := false
	done := false
	ip := p
	for ; p < len(s) && !done; p++ {
		switch s[p] {
		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08:
			return "", 0, fmt.Errorf("unexpected char ascii:%d in property param value", s[p])
		case 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B,
			0x1C, 0x1D, 0x1E, 0x1F:
			return "", 0, fmt.Errorf("unexpected char ascii:%d in property param value", s[p])
		case '\\':
			if p+2 >= len(s) {
				return "", 0, errors.New("unexpected end of param value")
			}
			r = append(r, []byte(FromText(string(s[p+1:p+2])))...)
			p++
			continue
		case ';', ':', ',':
			if !quoted {
				done = true
				p--
				continue
			}
		case '"':
			if p == ip {
				quoted = true
				continue
			}
			if quoted {
				done = true
				continue
			}
			return "", 0, fmt.Errorf("unexpected double quote in property param value")
		}
		r = append(r, s[p])
	}
	return string(r), p, nil
}

func parsePropertyValue(r *BaseProperty, contentLine string, p int) *BaseProperty {
	tokenPos := propertyValueTextReg.FindIndex([]byte(contentLine[p:]))
	if tokenPos == nil {
		return nil
	}
	r.Value = contentLine[p : p+tokenPos[1]]
	if r.GetValueType() == ValueDataTypeText {
		r.Value = FromText(r.Value)
	}
	return r
}

var textEscaper = strings.NewReplacer(
	`\`, `\\`,
	"\n", `\n`,
	`;`, `\;`,
	`,`, `\,`,
)

func ToText(s string) string {
	// Some special characters for iCalendar format should be escaped while
	// setting a value of a property with a TEXT type.
	return textEscaper.Replace(s)
}

var textUnescaper = strings.NewReplacer(
	`\\`, `\`,
	`\n`, "\n",
	`\N`, "\n",
	`\;`, `;`,
	`\,`, `,`,
)

func FromText(s string) string {
	// Some special characters for iCalendar format should be escaped while
	// setting a value of a property with a TEXT type.
	return textUnescaper.Replace(s)
}
