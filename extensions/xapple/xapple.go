package xapple

import (
	"errors"
	"image/color"

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
	cal.SetProperty(ical.Property(property), value, params...)
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

// SetCalendarColorFromColor sets the X-APPLE-CALENDAR-COLOR property from a color.Color.
func SetCalendarColorFromColor(cal *ical.Calendar, c color.Color, params ...ical.PropertyParameter) {
	SetCalendarColor(cal, string(ical.HexFromColor(c)), params...)
}

// GetCalendarColor returns the X-APPLE-CALENDAR-COLOR property from the calendar.
func GetCalendarColor(cal *ical.Calendar) *ical.CalendarProperty {
	if cal == nil {
		return nil
	}
	return cal.GetProperty(PropertyCalendarColor)
}

// GetCalendarColorAsString returns the X-APPLE-CALENDAR-COLOR value from the calendar.
func GetCalendarColorAsString(cal *ical.Calendar) string {
	p := GetCalendarColor(cal)
	if p == nil {
		return ""
	}
	return p.Value
}

// GetCalendarColorAsColor returns the X-APPLE-CALENDAR-COLOR value as a color.Color.
func GetCalendarColorAsColor(cal *ical.Calendar) (color.Color, error) {
	p := GetCalendarColor(cal)
	if p == nil {
		return nil, errors.New("x-apple-calendar-color property not found")
	}
	return ical.ColorFromHex(p.Value)
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
