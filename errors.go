package ics

const (
	MalformedCalendarExpectedVCalendarError  = "malformed calendar; expected a vcalendar"
	MalformedCalendarExpectedBeginError      = "malformed calendar; expected begin"
	MalformedCalendarExpectedEndError        = "malformed calendar; expected a end"
	MalformedCalendarExpectedBeginOrEndError = "malformed calendar; expected begin or end"

	MalformedCalendarUnexpectedEndError             = "malformed calendar; unexpected end"
	MalformedCalendarBadStateError                  = "malformed calendar; bad state"
	MalformedCalendarVCalendarNotWhereExpectedError = "malformed calendar; vcalendar not where expected"

	StartOrEndNotYetDefinedError = "start or end not yet defined"
	PropertyNotFoundError        = "property not found"
	ExpectedOneTZIDError         = "expected one TZID"

	TimeValueNotMatchedError                           = "time value not matched"
	TimeValueMatchedButUnsupportedAllDayTimeStampError = "time value matched but unsupported all-day timestamp"
	TimeValueMatchedButNotSupportedError               = "time value matched but not supported"

	ParsingComponentPropertyError = "parsing component property"
	ParsingComponentLineError     = "parsing component line"
	ParsingLineError              = "parsing line"
	ParsingCalendarLineError      = "parsing calendar line"
	ParsingPropertyError          = "parsing property"
	ParseError                    = "parse error"

	MissingPropertyValueError = "missing property value"

	UnexpectedASCIICharError                       = "unexpected char ascii"
	UnexpectedDoubleQuoteInPropertyParamValueError = "unexpected double quote in property param value"

	UnbalancedEndError = "unbalanced end"
	OutOfLinesError    = "ran out of lines"
)
