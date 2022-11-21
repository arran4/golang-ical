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
				v = strings.Replace(v, ";", "\\;", -1)
				v = strings.Replace(v, ":", "\\:", -1)
				v = strings.Replace(v, "\\", "\\\\", -1)
				v = strings.Replace(v, "\"", "\\\"", -1)
				v = strings.Replace(v, ",", "\\,", -1)
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
