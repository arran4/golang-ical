package ics

// The WithNewLine constants select the newline style used when serializing
// calendars.  RFC 5545 section 3.1 requires lines to be delimited by CRLF
// ("\r\n"), but many tools also accept LF on Unix systems.
const (
	// WithNewLineUnix uses LF line endings.
	WithNewLineUnix WithNewLine = "\n"
	// WithNewLineWindows uses CRLF line endings as required by RFC 5545 section 3.1.
	WithNewLineWindows WithNewLine = "\r\n"
)
