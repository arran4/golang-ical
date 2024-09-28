package ics

import "errors"

var (
	// ErrorPropertyNotFound is the error returned if the requested valid
	// property is not set.
	ErrorPropertyNotFound = errors.New("property not found")
)
