package ics

import "errors"

var (
	// ErrorPropertyNotFound is the error returned if the requested valid
	// property is not set.
	ErrorPropertyNotFound = errors.New("property not found")

	ErrorInvalidICalDuration                                      = errors.New("invalid iCal duration")
	ErrorInvalidICalDurationMissingDesignator                     = errors.New("invalid iCal duration: missing designator")
	ErrorInvalidICalDurationMissingTimeComponent                  = errors.New("invalid iCal duration: missing time component")
	ErrorInvalidICalDurationMissingValue                          = errors.New("invalid iCal duration: missing value")
	ErrorInvalidICalDurationDuplicateTimeDesignator               = errors.New("invalid iCal duration: duplicate time designator")
	ErrorInvalidICalDurationExpectedDigits                        = errors.New("invalid iCal duration: expected digits")
	ErrorInvalidICalDurationMissingUnit                           = errors.New("invalid iCal duration: missing unit")
	ErrorInvalidICalDurationWeeksOnlyDateComponent                = errors.New("invalid iCal duration: weeks must be the only date component")
	ErrorInvalidICalDurationWeeksOnlyComponent                    = errors.New("invalid iCal duration: weeks must be the only component")
	ErrorInvalidICalDurationDuplicateDayDesignator                = errors.New("invalid iCal duration: duplicate day designator")
	ErrorInvalidICalDurationHoursRequireTimeSection               = errors.New("invalid iCal duration: hours require a time section")
	ErrorInvalidICalDurationMinutesRequireTimeSection             = errors.New("invalid iCal duration: minutes require a time section")
	ErrorInvalidICalDurationSecondsRequireTimeSection             = errors.New("invalid iCal duration: seconds require a time section")
	ErrorInvalidICalDurationDuplicateOrOutOfOrderHoursComponent   = errors.New("invalid iCal duration: duplicate or out-of-order hours component")
	ErrorInvalidICalDurationDuplicateOrOutOfOrderMinutesComponent = errors.New("invalid iCal duration: duplicate or out-of-order minutes component")
	ErrorInvalidICalDurationDuplicateOrOutOfOrderSecondsComponent = errors.New("invalid iCal duration: duplicate or out-of-order seconds component")
	ErrorInvalidICalDurationUnknownUnit                           = errors.New("invalid iCal duration: unknown unit")
	// ErrInvalidOpArg marks an invalid variadic option argument.
	ErrInvalidOpArg = errors.New("invalid option argument")

	// ErrPropertySkipped marks a parser decision to skip a malformed line and continue.
	ErrPropertySkipped = errors.New("property skipped")
)
