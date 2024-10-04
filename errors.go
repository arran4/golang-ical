package ics

import (
	"errors"
	"fmt"
)

var (
	ErrUnexpectedParamValueLength = errors.New("unexpected end of param value")

	ErrMalformedCalendar = errors.New("malformed calendar")

	ErrMalformedCalendarExpectedVCalendar  = fmt.Errorf("%w: expected a vcalendar", ErrMalformedCalendar)
	ErrMalformedCalendarExpectedBegin      = fmt.Errorf("%w: expected begin", ErrMalformedCalendar)
	ErrMalformedCalendarExpectedEnd        = fmt.Errorf("%w: expected a end", ErrMalformedCalendar)
	ErrMalformedCalendarExpectedBeginOrEnd = fmt.Errorf("%w: expected begin or end", ErrMalformedCalendar)

	ErrMissingPropertyParamOperator               = fmt.Errorf("%w: missing property param operator", ErrMalformedCalendar)
	ErrUnexpectedEndOfProperty                    = fmt.Errorf("%w: unexpected end of property", ErrMalformedCalendar)
	ErrMalformedCalendarUnexpectedEnd             = fmt.Errorf("%w: unexpected end", ErrMalformedCalendar)
	ErrMalformedCalendarBadState                  = fmt.Errorf("%w: bad state", ErrMalformedCalendar)
	ErrMalformedCalendarVCalendarNotWhereExpected = fmt.Errorf("%w: vcalendar not where expected", ErrMalformedCalendar)

	ErrStartOrEndNotYetDefined = errors.New("start or end not yet defined")
	// ErrPropertyNotFound is the error returned if the requested valid
	// property is not set.
	ErrPropertyNotFound = errors.New("property not found")
	ErrExpectedOneTZID  = errors.New("expected one TZID")

	ErrTimeValueNotMatched                           = errors.New("time value not matched")
	ErrTimeValueMatchedButUnsupportedAllDayTimeStamp = errors.New("time value matched but unsupported all-day timestamp")
	ErrTimeValueMatchedButNotSupported               = errors.New("time value matched but not supported")

	ErrParsingComponentProperty = errors.New("parsing component property")
	ErrParsingComponentLine     = errors.New("parsing component line")
	ErrParsingLine              = errors.New("parsing line")
	ErrParsingCalendarLine      = errors.New("parsing calendar line")
	ErrParsingProperty          = errors.New("parsing property")
	ErrParse                    = errors.New("parse error")

	ErrMissingPropertyValue = errors.New("missing property value")

	ErrUnexpectedASCIIChar                       = errors.New("unexpected char ascii")
	ErrUnexpectedDoubleQuoteInPropertyParamValue = errors.New("unexpected double quote in property param value")

	ErrUnbalancedEnd = errors.New("unbalanced end")
	ErrOutOfLines    = errors.New("ran out of lines")
)
