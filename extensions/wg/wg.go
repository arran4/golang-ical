package wg

import (
	ical "github.com/arran4/golang-ical"
)

// PropertyTimezone is the X-WG-TIMEZONE property
const PropertyTimezone = "X-WG-TIMEZONE"

// PropertyCalName is the X-WG-CALNAME property
const PropertyCalName = "X-WG-CALNAME"

// PropertyCalDesc is the X-WG-CALDESC property
const PropertyCalDesc = "X-WG-CALDESC"

// PropertyRelCalId is the X-WG-RELCALID property
const PropertyRelCalId = "X-WG-RELCALID"

// SetProperty allows extending the properties easily
func SetProperty(cal *ical.Calendar, property string, value string, params ...ical.PropertyParameter) {
	if cal == nil {
		return
	}
	cal.SetProperty(ical.Property(property), value, params...)
}

// SetTimezone sets the X-WG-TIMEZONE property
func SetTimezone(cal *ical.Calendar, tz string, params ...ical.PropertyParameter) {
	SetProperty(cal, PropertyTimezone, tz, params...)
}

// SetCalName sets the X-WG-CALNAME property
func SetCalName(cal *ical.Calendar, name string, params ...ical.PropertyParameter) {
	SetProperty(cal, PropertyCalName, name, params...)
}

// SetCalDesc sets the X-WG-CALDESC property
func SetCalDesc(cal *ical.Calendar, desc string, params ...ical.PropertyParameter) {
	SetProperty(cal, PropertyCalDesc, desc, params...)
}

// SetRelCalId sets the X-WG-RELCALID property
func SetRelCalId(cal *ical.Calendar, id string, params ...ical.PropertyParameter) {
	SetProperty(cal, PropertyRelCalId, id, params...)
}
