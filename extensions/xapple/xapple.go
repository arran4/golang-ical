package xapple

import (
	ical "github.com/arran4/golang-ical"
)

const (
	// PropertyCalendarColor is the X-APPLE-CALENDAR-COLOR property
	PropertyCalendarColor ical.Property = "X-APPLE-CALENDAR-COLOR"

	// PropertyRegion is the X-APPLE-REGION property
	PropertyRegion ical.Property = "X-APPLE-REGION"
)

const (
	// ComponentPropertyStructuredLocation is the X-APPLE-STRUCTURED-LOCATION component property
	ComponentPropertyStructuredLocation ical.ComponentProperty = "X-APPLE-STRUCTURED-LOCATION"

	// ComponentPropertyTravelDuration is the X-APPLE-TRAVEL-DURATION component property
	ComponentPropertyTravelDuration ical.ComponentProperty = "X-APPLE-TRAVEL-DURATION"
)

type propertySetter interface {
	SetProperty(property ical.ComponentProperty, value string, props ...ical.PropertyParameter)
}

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

func SetComponentProperty(c ical.Component, property string, value string, params ...ical.PropertyParameter) {
	if c == nil {
		return
	}
	if ps, ok := c.(propertySetter); ok {
		ps.SetProperty(ical.ComponentProperty(property), value, params...)
	}
}

// Calendar properties

// SetCalendarColor sets the X-APPLE-CALENDAR-COLOR property for the calendar
func SetCalendarColor(cal *ical.Calendar, color string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyCalendarColor), color, params...)
}

// SetRegion sets the X-APPLE-REGION property for the calendar
func SetRegion(cal *ical.Calendar, region string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyRegion), region, params...)
}

// Component properties

// SetStructuredLocation sets the X-APPLE-STRUCTURED-LOCATION property for a component
func SetStructuredLocation(c ical.Component, location string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyStructuredLocation), location, params...)
}

// SetTravelDuration sets the X-APPLE-TRAVEL-DURATION property for a component
func SetTravelDuration(c ical.Component, duration string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyTravelDuration), duration, params...)
}
