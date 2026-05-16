package ics

import "errors"

var (
	// ErrorPropertyNotFound is the error returned if the requested valid
	// property is not set.
	ErrorPropertyNotFound = errors.New("property not found")

	ErrorInvalidICalDuration                                      = errors.New("invalid iCal duration")
	ErrorInvalidICalDurationMissingDesignator                     = errors.New("missing designator")
	ErrorInvalidICalDurationMissingTimeComponent                  = errors.New("missing time component")
	ErrorInvalidICalDurationMissingValue                          = errors.New("missing value")
	ErrorInvalidICalDurationDuplicateTimeDesignator               = errors.New("duplicate time designator")
	ErrorInvalidICalDurationExpectedDigits                        = errors.New("expected digits")
	ErrorInvalidICalDurationMissingUnit                           = errors.New("missing unit")
	ErrorInvalidICalDurationWeeksOnlyDateComponent                = errors.New("weeks must be the only date component")
	ErrorInvalidICalDurationWeeksOnlyComponent                    = errors.New("weeks must be the only component")
	ErrorInvalidICalDurationDuplicateDayDesignator                = errors.New("duplicate day designator")
	ErrorInvalidICalDurationHoursRequireTimeSection               = errors.New("hours require a time section")
	ErrorInvalidICalDurationMinutesRequireTimeSection             = errors.New("minutes require a time section")
	ErrorInvalidICalDurationSecondsRequireTimeSection             = errors.New("seconds require a time section")
	ErrorInvalidICalDurationDuplicateOrOutOfOrderHoursComponent   = errors.New("duplicate or out-of-order hours component")
	ErrorInvalidICalDurationDuplicateOrOutOfOrderMinutesComponent = errors.New("duplicate or out-of-order minutes component")
	ErrorInvalidICalDurationDuplicateOrOutOfOrderSecondsComponent = errors.New("duplicate or out-of-order seconds component")
	ErrorInvalidICalDurationUnknownUnit                           = errors.New("unknown unit")
	// ErrInvalidOpArg marks an invalid variadic option argument.
	ErrInvalidOpArg = errors.New("invalid option argument")

	// ErrPropertySkipped marks a parser decision to skip a malformed line and continue.
	ErrPropertySkipped = errors.New("property skipped")
)
