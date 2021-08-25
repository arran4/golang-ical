package ics

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
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

func (kv *KeyValues) KeyValue(s ...interface{}) (string, []string) {
	return kv.Key, kv.Value
}

func WithCN(cn string) PropertyParameter {
	return &KeyValues{
		Key:   string(ParameterCn),
		Value: []string{cn},
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
	lastSpace := -1
	for i, r := range s {
		if r == ' ' {
			lastSpace = i
		}

		newLength := length + utf8.RuneLen(r)
		if newLength > maxLength {
			break
		}
		length = newLength
	}
	if lastSpace > 0 {
		return s[:lastSpace]
	}

	return s[:length]
}

func (property *BaseProperty) serialize(w io.Writer) {
	b := bytes.NewBufferString("")
	fmt.Fprint(b, property.IANAToken)
	for k, vs := range property.ICalParameters {
		fmt.Fprint(b, ";")
		fmt.Fprint(b, k)
		fmt.Fprint(b, "=")
		for vi, v := range vs {
			if vi > 0 {
				fmt.Fprint(b, ",")
			}
			if strings.ContainsAny(v, ";:\\\",") {
				v = strings.Replace(v, "\"", "\\\"", -1)
				v = strings.Replace(v, "\\", "\\\\", -1)
			}
			fmt.Fprint(b, v)
		}
	}
	fmt.Fprint(b, ":")
	fmt.Fprint(b, property.Value)
	r := b.String()
	if len(r) > 75 {
		l := trimUT8StringUpTo(75, r)
		fmt.Fprint(w, l, "\r\n")
		r = r[len(l):]

		for len(r) > 74 {
			l := trimUT8StringUpTo(74, r)
			fmt.Fprint(w, " ", l, "\r\n")
			r = r[len(l):]
		}
		fmt.Fprint(w, " ")
	}
	fmt.Fprint(w, r, "\r\n")
}

type IANAProperty struct {
	BaseProperty
}

var (
	propertyIanaTokenReg  *regexp.Regexp
	propertyParamNameReg  *regexp.Regexp
	propertyParamValueReg *regexp.Regexp
	propertyValueTextReg  *regexp.Regexp
)

func init() {
	var err error
	propertyIanaTokenReg, err = regexp.Compile("[A-Za-z0-9-]{1,}")
	if err != nil {
		log.Panicf("Failed to build regex: %v", err)
	}
	propertyParamNameReg = propertyIanaTokenReg
	propertyParamValueReg, err = regexp.Compile("^(?:\"(?:[^\"\\\\]|\\[\"nrt])*\"|[^,;\\\\:\"]*)")
	if err != nil {
		log.Panicf("Failed to build regex: %v", err)
	}
	propertyValueTextReg, err = regexp.Compile("^.*")
	if err != nil {
		log.Panicf("Failed to build regex: %v", err)
	}
}

type ContentLine string

func ParseProperty(contentLine ContentLine) *BaseProperty {
	r := &BaseProperty{
		ICalParameters: map[string][]string{},
	}
	tokenPos := propertyIanaTokenReg.FindIndex([]byte(contentLine))
	if tokenPos == nil {
		return nil
	}
	p := 0
	r.IANAToken = string(contentLine[p+tokenPos[0] : p+tokenPos[1]])
	p += tokenPos[1]
	for {
		if p >= len(contentLine) {
			return nil
		}
		switch rune(contentLine[p]) {
		case ':':
			return parsePropertyValue(r, string(contentLine), p+1)
		case ';':
			var np int
			r, np = parsePropertyParam(r, string(contentLine), p+1)
			if r == nil {
				return nil
			}
			p = np
		default:
			return nil
		}
	}
}

func parsePropertyParam(r *BaseProperty, contentLine string, p int) (*BaseProperty, int) {
	tokenPos := propertyParamNameReg.FindIndex([]byte(contentLine[p:]))
	if tokenPos == nil {
		return nil, p
	}
	k, v := "", ""
	k = string(contentLine[p : p+tokenPos[1]])
	p += tokenPos[1]
	switch rune(contentLine[p]) {
	case '=':
		p += 1
	default:
		return nil, p
	}
	for {
		if p >= len(contentLine) {
			return nil, p
		}
		tokenPos = propertyParamValueReg.FindIndex([]byte(contentLine[p:]))
		if tokenPos == nil {
			return nil, p
		}
		v = string(contentLine[p+tokenPos[0] : p+tokenPos[1]])
		p += tokenPos[1]
		r.ICalParameters[k] = append(r.ICalParameters[k], v)
		switch rune(contentLine[p]) {
		case ',':
			p += 1
		default:
			return r, p
		}
	}
}

func parsePropertyValue(r *BaseProperty, contentLine string, p int) *BaseProperty {
	tokenPos := propertyValueTextReg.FindIndex([]byte(contentLine[p:]))
	if tokenPos == nil {
		return nil
	}
	r.Value = string(contentLine[p : p+tokenPos[1]])
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
