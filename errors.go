package ics

import "errors"

var (
	ErrorUnsupportedOptionalArgument = errors.New("unsupported time property type")

	// ErrorPropertyNotFound is the error returned if the requested valid
	// property is not set.
	ErrorPropertyNotFound = errors.New("property not found")

	// ErrInvalidOpArg marks an invalid variadic option argument.
	ErrInvalidOpArg = errors.New("invalid option argument")

	// ErrPropertySkipped marks a parser decision to skip a malformed line and continue.
	ErrPropertySkipped = errors.New("property skipped")
)
