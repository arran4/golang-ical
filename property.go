package ics

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

func (bp *BaseProperty) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	var b strings.Builder
	b.WriteString(bp.IANAToken)

	var keys []string
	for k := range bp.ICalParameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := bp.ICalParameters[k]
		b.WriteByte(';')
		b.WriteString(k)
		b.WriteByte('=')
		for vi, v := range vs {
			if vi > 0 {
				b.WriteByte(',')
			}
			if Parameter(k).IsQuoted() {
				v = quotedValueString(v)
				b.WriteString(v)
			} else {
				v = escapeValueString(v)
				b.WriteString(v)
			}
		}
	}
	b.WriteByte(':')
	propertyValue := bp.Value
	if bp.GetValueType() == ValueDataTypeText {
		propertyValue = ToText(propertyValue)
	}
	b.WriteString(propertyValue)
	r := b.String()
	if len(r) > serialConfig.MaxLength {
		l := trimUT8StringUpTo(serialConfig.MaxLength, r)
		_, err := io.WriteString(w, l+serialConfig.NewLine)
		if err != nil {
			return fmt.Errorf("property %s serialization: %w", bp.IANAToken, err)
		}
		r = r[len(l):]

		for len(r) > serialConfig.MaxLength-1 {
			l := trimUT8StringUpTo(serialConfig.MaxLength-1, r)
			_, err = io.WriteString(w, " "+l+serialConfig.NewLine)
			if err != nil {
				return fmt.Errorf("property %s serialization: %w", bp.IANAToken, err)
			}
			r = r[len(l):]
		}
		_, err = io.WriteString(w, " ")
		if err != nil {
			return fmt.Errorf("property %s serialization: %w", bp.IANAToken, err)
		}
	}
	_, err := io.WriteString(w, r+serialConfig.NewLine)
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

// PropertyParser is an optional replacement parser for malformed content lines.
// It receives the raw content line and can either recover, skip, or abort.
type PropertyParser func(rawLine ContentLine) (*BaseProperty, error)

// ParseProperty parses a single RFC5545 content line using strict parsing rules.
func ParseProperty(contentLine ContentLine) (*BaseProperty, error) {
	return parseProperty(contentLine)
}

// SkipPropertyParser is an alternative to the default strict parser path used by ParseCalendar and ParseComponent.
// It returns parsed properties when the line is valid and skips malformed lines.
func SkipPropertyParser(rawLine ContentLine) (*BaseProperty, error) {
	line, err := parseProperty(rawLine)
	if err != nil {
		return nil, nil
	}
	return line, nil
}

// LooseParser is an alternative to the default strict parser path used by ParseCalendar and ParseComponent.
// It preserves the property token and value when malformed parameters prevent a full parse.
func LooseParser(rawLine ContentLine) (*BaseProperty, error) {
	s := string(rawLine)
	colonIdx := strings.Index(s, ":")
	if colonIdx <= 0 {
		return nil, fmt.Errorf("%w: unable to recover property (no colon)", ErrPropertySkipped)
	}
	tokenEnd := colonIdx
	if semiIdx := strings.Index(s, ";"); semiIdx > 0 && semiIdx < colonIdx {
		tokenEnd = semiIdx
	}
	return &BaseProperty{
		IANAToken:      s[:tokenEnd],
		Value:          s[colonIdx+1:],
		ICalParameters: map[string][]string{},
	}, nil
}

// FallbackParser builds an alternative to the default strict parser path used by ParseCalendar and ParseComponent.
// It tries the strict parser first, then each supplied fallback parser in order.
func FallbackParser(fallback PropertyParser, fallbacks ...PropertyParser) PropertyParser {
	parsers := append([]PropertyParser{fallback}, fallbacks...)
	return func(rawLine ContentLine) (*BaseProperty, error) {
		line, err := parseProperty(rawLine)
		if err == nil {
			return line, nil
		}
		for _, p := range parsers {
			if p == nil {
				continue
			}
			line, err := p(rawLine)
			if err != nil {
				if errors.Is(err, ErrPropertySkipped) {
					continue
				}
				return nil, err
			}
			if line != nil {
				return line, nil
			}
		}
		return nil, fmt.Errorf("%w: no capable parsers found", ErrPropertySkipped)
	}
}

func parsePropertyParserOptions(current PropertyParser, opts ...any) (PropertyParser, error) {
	for i, opt := range opts {
		switch opt := opt.(type) {
		case PropertyParser:
			if opt != nil {
				current = opt
			}
		case func(ContentLine) (*BaseProperty, error):
			if opt != nil {
				current = PropertyParser(opt)
			}
		default:
			return current, fmt.Errorf("%w %d: %T", ErrInvalidOpArg, i, opt)
		}
	}
	return current, nil
}

func parseProperty(contentLine ContentLine) (*BaseProperty, error) {
	r := &BaseProperty{
		ICalParameters: map[string][]string{},
	}
	tokenPos := propertyIanaTokenReg.FindIndex([]byte(contentLine))
	if tokenPos == nil {
		return nil, fmt.Errorf("invalid property token in %q", contentLine)
	}
	p := 0
	r.IANAToken = string(contentLine[p+tokenPos[0] : p+tokenPos[1]])
	p += tokenPos[1]
	for {
		if p >= len(contentLine) {
			return nil, fmt.Errorf("unexpected end of property %s", r.IANAToken)
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
				return nil, fmt.Errorf("%w %s: %w", ErrParsingProperty, t, err)
			}
			if r == nil {
				return nil, fmt.Errorf("parsing property %s: invalid property", t)
			}
			p = np
		default:
			return nil, fmt.Errorf("parsing property %s: unexpected character %q", r.IANAToken, contentLine[p])
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
		return nil, p, fmt.Errorf("%w for %s in %s", ErrMissingPropertyParamOperator, k, r.IANAToken)
	}
	switch rune(contentLine[p]) {
	case '=':
		p += 1
	default:
		return nil, p, fmt.Errorf("%w for %s in %s", ErrMissingPropertyValue, k, r.IANAToken)
	}
	for {
		if p >= len(contentLine) {
			return nil, p, nil
		}
		var err error
		v, p, err = parsePropertyParamValue(contentLine, p)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: %w %s in %s", ErrParse, err, k, r.IANAToken)
		}
		r.ICalParameters[k] = append(r.ICalParameters[k], v)
		if p >= len(contentLine) {
			return nil, p, fmt.Errorf("%w %s", ErrUnexpectedEndOfProperty, r.IANAToken)
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
			return "", 0, fmt.Errorf("%w:%d in property param value", ErrUnexpectedASCIIChar, s[p])
		case 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B,
			0x1C, 0x1D, 0x1E, 0x1F:
			return "", 0, fmt.Errorf("%w:%d in property param value", ErrUnexpectedASCIIChar, s[p])
		case '\\':
			if p+2 >= len(s) {
				return "", 0, ErrUnexpectedParamValueLength
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
			return "", 0, ErrUnexpectedDoubleQuoteInPropertyParamValue
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
