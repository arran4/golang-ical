package wr

import (
	ical "github.com/arran4/golang-ical"
)

// PropertyTimezone is the X-WR-TIMEZONE property
const PropertyTimezone = "X-WR-TIMEZONE"

// PropertyCalName is the X-WR-CALNAME property
const PropertyCalName = "X-WR-CALNAME"

// PropertyCalDesc is the X-WR-CALDESC property
const PropertyCalDesc = "X-WR-CALDESC"

// PropertyRelCalId is the X-WR-RELCALID property
const PropertyRelCalId = "X-WR-RELCALID"

// SetProperty allows extending the properties easily
func SetProperty(cal *ical.Calendar, property string, value string, params ...ical.PropertyParameter) {
	if cal == nil {
		return
	}
	// Fallback to directly modify CalendarProperties if no public SetProperty exists
	found := false
	for i := range cal.CalendarProperties {
		if cal.CalendarProperties[i].IANAToken == property {
			cal.CalendarProperties[i].Value = value
			cal.CalendarProperties[i].ICalParameters = map[string][]string{}
			for _, p := range params {
				k, v := p.KeyValue()
				cal.CalendarProperties[i].ICalParameters[k] = v
			}
			found = true
			break
		}
	}
	if !found {
		r := ical.CalendarProperty{
			BaseProperty: ical.BaseProperty{
				IANAToken:      property,
				Value:          value,
				ICalParameters: map[string][]string{},
			},
		}
		for _, p := range params {
			k, v := p.KeyValue()
			r.ICalParameters[k] = v
		}
		cal.CalendarProperties = append(cal.CalendarProperties, r)
	}
}

// SetTimezone sets the X-WR-TIMEZONE property
func SetTimezone(cal *ical.Calendar, tz string, params ...ical.PropertyParameter) {
	if cal == nil {
		return
	}
	cal.SetXWRTimezone(tz, params...)
}

// SetCalName sets the X-WR-CALNAME property
func SetCalName(cal *ical.Calendar, name string, params ...ical.PropertyParameter) {
	if cal == nil {
		return
	}
	cal.SetXWRCalName(name, params...)
}

// SetCalDesc sets the X-WR-CALDESC property
func SetCalDesc(cal *ical.Calendar, desc string, params ...ical.PropertyParameter) {
	if cal == nil {
		return
	}
	cal.SetXWRCalDesc(desc, params...)
}

// SetRelCalId sets the X-WR-RELCALID property
func SetRelCalId(cal *ical.Calendar, id string, params ...ical.PropertyParameter) {
	if cal == nil {
		return
	}
	cal.SetXWRCalID(id, params...)
}
