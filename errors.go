package ics

import (
	"errors"
)

var (
	ErrStartAndEndDateNotDefined = errors.New("start time and end time not defined")
	// ErrorPropertyNotFound is the error returned if the requested valid
	// property is not set.
	ErrorPropertyNotFound = errors.New("property not found")
)
