package ics

// NewLineString defines the default newline for Windows systems.  It resolves
// to WithNewLineWindows which uses CRLF line endings as mandated by RFC 5545.
const (
	NewLineString = WithNewLineWindows
)
