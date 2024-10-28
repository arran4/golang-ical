package ics

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Component To determine what this is please use a type switch or typecast to each of:
// - *VEvent
// - *VTodo
// - *VBusy
// - *VJournal
type Component interface {
	UnknownPropertiesIANAProperties() []IANAProperty
	SubComponents() []Component
	SerializeTo(b io.Writer, serialConfig *SerializationConfiguration) error
}

var (
	_ Component = (*VEvent)(nil)
	_ Component = (*VTodo)(nil)
	_ Component = (*VBusy)(nil)
	_ Component = (*VJournal)(nil)
)

type ComponentBase struct {
	Properties []IANAProperty
	Components []Component
}

func (cb *ComponentBase) UnknownPropertiesIANAProperties() []IANAProperty {
	return cb.Properties
}

func (cb *ComponentBase) SubComponents() []Component {
	return cb.Components
}

func (cb *ComponentBase) serializeThis(writer io.Writer, componentType ComponentType, serialConfig *SerializationConfiguration) error {
	_, _ = fmt.Fprint(writer, "BEGIN:"+componentType, serialConfig.NewLine)
	for _, p := range cb.Properties {
		err := p.serialize(writer, serialConfig)
		if err != nil {
			return err
		}
	}
	for _, c := range cb.Components {
		err := c.SerializeTo(writer, serialConfig)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(writer, "END:"+componentType, serialConfig.NewLine)
	return err
}

func NewComponent(uniqueId string) ComponentBase {
	return ComponentBase{
		Properties: []IANAProperty{
			{BaseProperty{IANAToken: string(ComponentPropertyUniqueId), Value: uniqueId}},
		},
	}
}

// GetProperty returns the first match for the particular property you're after. Please consider using:
// ComponentProperty.Required to determine if GetProperty or GetProperties is more appropriate.
func (cb *ComponentBase) GetProperty(componentProperty ComponentProperty) *IANAProperty {
	return IANAProperties(cb.Properties).GetProperty(componentProperty)
}

// GetProperties returns all matches for the particular property you're after. Please consider using:
// ComponentProperty.Singular/ComponentProperty.Multiple to determine if GetProperty or GetProperties is more appropriate.
func (cb *ComponentBase) GetProperties(componentProperty ComponentProperty) []*IANAProperty {
	var result []*IANAProperty
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken == string(componentProperty) {
			result = append(result, &cb.Properties[i])
		}
	}
	return result
}

// HasProperty returns true if a component property is in the component.
func (cb *ComponentBase) HasProperty(componentProperty ComponentProperty) bool {
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken == string(componentProperty) {
			return true
		}
	}
	return false
}

// SetProperty replaces the first match for the particular property you're setting, otherwise adds it. Please consider using:
// ComponentProperty.Singular/ComponentProperty.Multiple to determine if AddProperty, SetProperty or ReplaceProperty is
// more appropriate.
func (cb *ComponentBase) SetProperty(property ComponentProperty, value string, params ...PropertyParameter) {
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken == string(property) {
			cb.Properties[i].Value = value
			cb.Properties[i].ICalParameters = map[string][]string{}
			for _, p := range params {
				k, v := p.KeyValue()
				cb.Properties[i].ICalParameters[k] = v
			}
			return
		}
	}
	cb.AddProperty(property, value, params...)
}

// ReplaceProperty replaces all matches of the particular property you're setting, otherwise adds it. Returns a slice
// of removed properties. Please consider using:
// ComponentProperty.Singular/ComponentProperty.Multiple to determine if AddProperty, SetProperty or ReplaceProperty is
// more appropriate.
func (cb *ComponentBase) ReplaceProperty(property ComponentProperty, value string, params ...PropertyParameter) []IANAProperty {
	removed := cb.RemoveProperty(property)
	cb.AddProperty(property, value, params...)
	return removed
}

// AddProperty appends a property
func (cb *ComponentBase) AddProperty(property ComponentProperty, value string, params ...PropertyParameter) {
	r := IANAProperty{
		BaseProperty{
			IANAToken:      string(property),
			Value:          value,
			ICalParameters: map[string][]string{},
		},
	}
	for _, p := range params {
		k, v := p.KeyValue()
		r.ICalParameters[k] = v
	}
	cb.Properties = append(cb.Properties, r)
}

// RemoveProperty removes from the component all properties that is of a particular property type, returning an slice of
// removed entities
func (cb *ComponentBase) RemoveProperty(removeProp ComponentProperty) []IANAProperty {
	var keptProperties []IANAProperty
	var removedProperties []IANAProperty
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken != string(removeProp) {
			keptProperties = append(keptProperties, cb.Properties[i])
		} else {
			removedProperties = append(removedProperties, cb.Properties[i])
		}
	}
	cb.Properties = keptProperties
	return removedProperties
}

// RemovePropertyByValue removes from the component all properties that has a particular property type and value,
// return a count of removed properties
func (cb *ComponentBase) RemovePropertyByValue(removeProp ComponentProperty, value string) []IANAProperty {
	return cb.RemovePropertyByFunc(removeProp, func(p IANAProperty) bool {
		return p.Value == value
	})
}

// RemovePropertyByFunc removes from the component all properties that has a particular property type and the function
// remove returns true for
func (cb *ComponentBase) RemovePropertyByFunc(removeProp ComponentProperty, remove func(p IANAProperty) bool) []IANAProperty {
	var keptProperties []IANAProperty
	var removedProperties []IANAProperty
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken != string(removeProp) && remove(cb.Properties[i]) {
			keptProperties = append(keptProperties, cb.Properties[i])
		} else {
			removedProperties = append(removedProperties, cb.Properties[i])
		}
	}
	cb.Properties = keptProperties
	return removedProperties
}

const (
	icalTimestampFormatUTC = "20060102T150405Z"
	icalTimestampFormat    = "20060102T150405"
	icalDateFormatUTC      = "20060102Z"
	icalDateFormat         = "20060102"
)

func dateFormatForTime(t time.Time, props []PropertyParameter) (string, []PropertyParameter) {
	// For implementation details see legacy + https://icalendar.org/iCalendar-RFC-5545/3-3-5-date-time.html
	layout := icalDateFormat
	tzid := PropertyParameters(props).GetProperty(ComponentPropertyTzid)
	l := t.Location()
	var ls string
	if l != nil {
		ls = l.String()
	}
	if (l == time.UTC /* || ls == "MST"*/ || ls == "") && tzid == nil {
		layout = icalDateFormatUTC
	} else if tzid == nil && l != time.Local /* && ls != "MST"*/ {
		props = append(props, WithTZID(ls))
	}
	return layout, props
}

func timestampFormatForTime(t time.Time, props []PropertyParameter) (string, []PropertyParameter) {
	layout := icalTimestampFormat
	tzid := PropertyParameters(props).GetProperty(ComponentPropertyTzid)
	l := t.Location()
	var ls string
	if l != nil {
		ls = l.String()
	}
	if (l == time.UTC || /*ls == "MST" ||*/ ls == "") && tzid == nil {
		layout = icalTimestampFormatUTC
	} else if tzid == nil && l != time.Local /* && ls != "MST"*/ {
		props = append(props, WithTZID(ls))
	}
	return layout, props
}

/*
 * RFC 2445                       iCalendar                   November 1998
 *
 * 4.3.4 Date
 *
 * Value Name: DATE
 *
 * Purpose: This value type is used to identify values that contain a
 * calendar date.
 *
 * Formal Definition: The value type is defined by the following
 * notation:
 *
 *   date               = date-value
 *
 *   date-value         = date-fullyear date-month date-mday
 *   date-fullyear      = 4DIGIT
 *
 *  date-month         = 2DIGIT        ;01-12
 *  date-mday          = 2DIGIT        ;01-28, 01-29, 01-30, 01-31
 * 									;based on month/year
 *
 * Description: If the property permits, multiple "date" values are
 * specified as a COMMA character (US-ASCII decimal 44) separated list
 * of values. The format for the value type is expressed as the [ISO
 * 8601] complete representation, basic format for a calendar date. The
 * textual format specifies a four-digit year, two-digit month, and
 * two-digit day of the month. There are no separator characters between
 * the year, month and day component text.
 *
 * No additional content value encoding (i.e., BACKSLASH character
 * encoding) is defined for this value type.
 *
 * Example: The following represents July 14, 1997:
 *
 *  19970714
 *
 * Value Name: DATE-TIME
 *
 * Purpose: This value type is used to identify values that specify a
 * precise calendar date and time of day.
 *
 * Formal Definition: The value type is defined by the following
 * notation:
 *
 *  date-time  = date "T" time ;As specified in the date and time
 * 							;value definitions
 *
 * Description: If the property permits, multiple "date-time" values are
 * specified as a COMMA character (US-ASCII decimal 44) separated list
 * of values. No additional content value encoding (i.e., BACKSLASH
 * character encoding) is defined for this value type.
 *
 * The "DATE-TIME" data type is used to identify values that contain a
 * precise calendar date and time of day. The format is based on the
 * [ISO 8601] complete representation, basic format for a calendar date
 * and time of day. The text format is a concatenation of the "date",
 * followed by the LATIN CAPITAL LETTER T character (US-ASCII decimal
 * 84) time designator, followed by the "time" format.
 *
 * The "DATE-TIME" data type expresses time values in three forms:
 *
 * The form of date and time with UTC offset MUST NOT be used. For
 * example, the following is not valid for a date-time value:
 *
 *  DTSTART:19980119T230000-0800       ;Invalid time format
 *
 * FORM #1: DATE WITH LOCAL TIME
 *
 * The date with local time form is simply a date-time value that does
 * not contain the UTC designator nor does it reference a time zone. For
 * example, the following represents Janurary 18, 1998, at 11 PM:
 *
 *  DTSTART:19980118T230000
 *
 * Date-time values of this type are said to be "floating" and are not
 * bound to any time zone in particular. They are used to represent the
 * same hour, minute, and second value regardless of which time zone is
 * currently being observed. For example, an event can be defined that
 * indicates that an individual will be busy from 11:00 AM to 1:00 PM
 * every day, no matter which time zone the person is in. In these
 * cases, a local time can be specified. The recipient of an iCalendar
 * object with a property value consisting of a local time, without any
 * relative time zone information, SHOULD interpret the value as being
 * fixed to whatever time zone the ATTENDEE is in at any given moment.
 * This means that two ATTENDEEs, in different time zones, receiving the
 * same event definition as a floating time, may be participating in the
 * event at different actual times. Floating time SHOULD only be used
 * where that is the reasonable behavior.
 *
 * In most cases, a fixed time is desired. To properly communicate a
 * fixed time in a property value, either UTC time or local time with
 * time zone reference MUST be specified.
 *
 * The use of local time in a DATE-TIME value without the TZID property
 * parameter is to be interpreted as floating time, regardless of the
 * existence of "VTIMEZONE" calendar components in the iCalendar object.
 *
 * FORM #2: DATE WITH UTC TIME
 *
 * The date with UTC time, or absolute time, is identified by a LATIN
 * CAPITAL LETTER Z suffix character (US-ASCII decimal 90), the UTC
 * designator, appended to the time value. For example, the following
 * represents January 19, 1998, at 0700 UTC:
 *
 *  DTSTART:19980119T070000Z
 *
 * The TZID property parameter MUST NOT be applied to DATE-TIME
 * properties whose time values are specified in UTC.
 *
 * FORM #3: DATE WITH LOCAL TIME AND TIME ZONE REFERENCE
 *
 * The date and local time with reference to time zone information is
 * identified by the use the TZID property parameter to reference the
 * appropriate time zone definition. TZID is discussed in detail in the
 * section on Time Zone. For example, the following represents 2 AM in
 * New York on Janurary 19, 1998:
 *
 * 	  DTSTART;TZID=US-Eastern:19980119T020000
 *
 * Example: The following represents July 14, 1997, at 1:30 PM in New
 * York City in each of the three time formats, using the "DTSTART"
 * property.
 *
 *  DTSTART:19970714T133000            ;Local time
 *  DTSTART:19970714T173000Z           ;UTC time
 *  DTSTART;TZID=US-Eastern:19970714T133000    ;Local time and time
 * 					; zone reference
 *
 * A time value MUST ONLY specify 60 seconds when specifying the
 * periodic "leap second" in the time value. For example:
 *
 *  COMPLETED:19970630T235960Z
 */
var timeStampVariations = regexp.MustCompile("^([0-9]{8})?([TZ])?([0-9]{6})?(Z)?$")

func (cb *ComponentBase) SetCreatedTime(t time.Time, params ...PropertyParameter) {
	layout, params := timestampFormatForTime(t, params)
	cb.SetProperty(ComponentPropertyCreated, t.Format(layout), params...)
}

func (cb *ComponentBase) SetDtStampTime(t time.Time, params ...PropertyParameter) {
	layout, params := timestampFormatForTime(t, params)
	cb.SetProperty(ComponentPropertyDtstamp, t.Format(layout), params...)
}

func (cb *ComponentBase) SetModifiedAt(t time.Time, params ...PropertyParameter) {
	layout, params := timestampFormatForTime(t, params)
	cb.SetProperty(ComponentPropertyLastModified, t.Format(layout), params...)
}

func (cb *ComponentBase) SetSequence(seq int, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertySequence, strconv.Itoa(seq), params...)
}

func (cb *ComponentBase) SetStartAt(t time.Time, params ...PropertyParameter) {
	layout, params := timestampFormatForTime(t, params)
	cb.SetProperty(ComponentPropertyDtStart, t.Format(layout), params...)
}

func (cb *ComponentBase) SetAllDayStartAt(t time.Time, params ...PropertyParameter) {
	layout, params := dateFormatForTime(t, params)
	cb.SetProperty(
		ComponentPropertyDtStart,
		t.Format(layout),
		append(params, WithValue(string(ValueDataTypeDate)))...,
	)
}

func (cb *ComponentBase) SetEndAt(t time.Time, params ...PropertyParameter) {
	layout, params := timestampFormatForTime(t, params)
	cb.SetProperty(ComponentPropertyDtEnd, t.Format(layout), params...)
}

func (cb *ComponentBase) SetAllDayEndAt(t time.Time, params ...PropertyParameter) {
	layout, params := dateFormatForTime(t, params)
	cb.SetProperty(
		ComponentPropertyDtEnd,
		t.Format(layout),
		append(params, WithValue(string(ValueDataTypeDate)))...,
	)
}

// SetDuration updates the duration of an event.
// This function will set either the end or start time of an event depending on what is already given.
// The duration defines the length of an event relative to start or end time.
//
// Notice: It will not set the DURATION key of the ics - only DTSTART and DTEND will be affected. See SetDurationStr for
// setting the actual property.
func (cb *ComponentBase) SetDuration(d time.Duration) error {
	return cb.SetDurationWithParameters(d)
}

// SetDurationWithParameters is the same as SetDuration but with parameters which as passed to the appropriate 'set'
func (cb *ComponentBase) SetDurationWithParameters(d time.Duration, params ...PropertyParameter) error {
	startProp := cb.GetProperty(ComponentPropertyDtStart)
	if startProp != nil {
		t, allDay, err := startProp.ParseTime(false)
		if t != nil && err == nil {
			v, _ := startProp.parameterValue(ParameterValue)
			if v == string(ValueDataTypeDate) || allDay {
				cb.SetAllDayEndAt(t.Add(d), params...)
			} else {
				cb.SetEndAt(t.Add(d), params...)
			}
			return nil
		}
	}
	endProp := cb.GetProperty(ComponentPropertyDtEnd)
	if endProp != nil {
		t, allDay, err := endProp.ParseTime(false)
		if t != nil && err == nil {
			v, _ := endProp.parameterValue(ParameterValue)
			if v == string(ValueDataTypeDate) || allDay {
				cb.SetAllDayStartAt(t.Add(-d), params...)
			} else {
				cb.SetStartAt(t.Add(-d), params...)
			}
			return nil
		}
	}
	return errors.New("start or end not yet defined")
}

func (cb *ComponentBase) IsDuring(point time.Time, ops ...any) (bool, error) {
	var effectiveStartTime *time.Time
	var effectiveEndTime *time.Time
	var durations []Duration
	var startAllDay bool
	var endAllDay bool
	var err error
	startProp := cb.GetProperty(ComponentPropertyDtStart)
	if startProp != nil {
		effectiveStartTime, startAllDay, err = startProp.ParseTime(false)
		if err != nil {
			return false, fmt.Errorf("start time: %w", err)
		}
	}
	endProp := cb.GetProperty(ComponentPropertyDtEnd)
	if endProp != nil {
		effectiveEndTime, endAllDay, err = endProp.ParseTime(false)
		if err != nil {
			return false, fmt.Errorf("start time: %w", err)
		}
	}
	durationProp := cb.GetProperty(ComponentPropertyDuration)
	if durationProp != nil {
		durations, err = durationProp.ParseDurations()
		if err != nil {
			return false, fmt.Errorf("start time: %w", err)
		}
	}
	switch {
	case len(durations) == 1 && effectiveStartTime == nil && effectiveEndTime != nil:
		d := durations[0].Duration
		days := durations[0].Days
		// TODO clarify expected behavior
		if durations[0].Positive {
			d = -d
			days = -days
		}
		t := effectiveEndTime.Add(d).AddDate(0, 0, days)
		effectiveStartTime = &t
	case len(durations) == 1 && effectiveStartTime != nil && effectiveEndTime == nil:
		d := durations[0].Duration
		days := durations[0].Days
		// TODO clarify expected behavior
		if !durations[0].Positive {
			d = -d
			days = -days
		}
		t := effectiveStartTime.Add(d).AddDate(0, 0, days)
		effectiveEndTime = &t
	case effectiveStartTime == nil && effectiveEndTime == nil:
		return false, ErrStartAndEndDateNotDefined
	}
	if effectiveStartTime != nil && effectiveEndTime != nil {
		// If it starts and ends on the same day and at least one of the "allDays" is set, then it is an all day
		if effectiveStartTime.Truncate(24*time.Hour-1).Equal(effectiveEndTime.Truncate(24*time.Hour-1)) && (startAllDay || endAllDay) {
			startAllDay, endAllDay = true, true
		}
	}
	if startAllDay && effectiveStartTime != nil {
		t := effectiveStartTime.Truncate(24*time.Hour + 1)
		effectiveStartTime = &t
	}
	if endAllDay && effectiveEndTime != nil {
		t := effectiveEndTime.AddDate(0, 0, 1).Truncate(24*time.Hour - 1).Add(-1)
		effectiveEndTime = &t
	}
	switch {
	case effectiveStartTime == nil && effectiveEndTime == nil:
		return false, nil
	case effectiveStartTime != nil && effectiveEndTime != nil:
		return (point.Equal(*effectiveStartTime) || point.After(*effectiveStartTime)) && (point.Equal(*effectiveEndTime) || point.Before(*effectiveEndTime)), nil
	}
	return false, fmt.Errorf("unsupported state")
}

func (cb *ComponentBase) GetEndAt() (time.Time, error) {
	return cb.getTimeProp(ComponentPropertyDtEnd, false)
}

func (cb *ComponentBase) getTimeProp(componentProperty ComponentProperty, expectAllDay bool) (time.Time, error) {
	timeProp := cb.GetProperty(componentProperty)
	if timeProp == nil {
		return time.Time{}, fmt.Errorf("%w: %s", ErrorPropertyNotFound, componentProperty)
	}
	t, _, err := timeProp.ParseTime(expectAllDay)
	if t == nil {
		return time.Time{}, err
	}
	return *t, err
}

func (cb *ComponentBase) GetStartAt() (time.Time, error) {
	return cb.getTimeProp(ComponentPropertyDtStart, false)
}

func (cb *ComponentBase) GetAllDayStartAt() (time.Time, error) {
	return cb.getTimeProp(ComponentPropertyDtStart, true)
}

func (cb *ComponentBase) GetLastModifiedAt() (time.Time, error) {
	return cb.getTimeProp(ComponentPropertyLastModified, false)
}

func (cb *ComponentBase) GetDtStampTime() (time.Time, error) {
	return cb.getTimeProp(ComponentPropertyDtstamp, false)
}

func (cb *ComponentBase) SetSummary(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertySummary, s, params...)
}

func (cb *ComponentBase) SetStatus(s ObjectStatus, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyStatus, string(s), params...)
}

func (cb *ComponentBase) SetDescription(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDescription, s, params...)
}

func (cb *ComponentBase) SetLocation(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyLocation, s, params...)
}

func (cb *ComponentBase) setGeo(lat interface{}, lng interface{}, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyGeo, fmt.Sprintf("%v;%v", lat, lng), params...)
}

func (cb *ComponentBase) SetURL(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyUrl, s, params...)
}

func (cb *ComponentBase) SetOrganizer(s string, params ...PropertyParameter) {
	if !strings.HasPrefix(s, "mailto:") {
		s = "mailto:" + s
	}

	cb.SetProperty(ComponentPropertyOrganizer, s, params...)
}

func (cb *ComponentBase) SetColor(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyColor, s, params...)
}

func (cb *ComponentBase) SetClass(c Classification, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyClass, string(c), params...)
}

func (cb *ComponentBase) setPriority(p int, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyPriority, strconv.Itoa(p), params...)
}

func (cb *ComponentBase) setResources(r string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyResources, r, params...)
}

func (cb *ComponentBase) AddAttendee(s string, params ...PropertyParameter) {
	if !strings.HasPrefix(s, "mailto:") {
		s = "mailto:" + s
	}

	cb.AddProperty(ComponentPropertyAttendee, s, params...)
}

func (cb *ComponentBase) AddExdate(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyExdate, s, params...)
}

func (cb *ComponentBase) AddExrule(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyExrule, s, params...)
}

func (cb *ComponentBase) AddRdate(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyRdate, s, params...)
}

func (cb *ComponentBase) AddRrule(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyRrule, s, params...)
}

func (cb *ComponentBase) AddAttachment(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyAttach, s, params...)
}

func (cb *ComponentBase) AddAttachmentURL(uri string, contentType string) {
	cb.AddAttachment(uri, WithFmtType(contentType))
}

func (cb *ComponentBase) AddAttachmentBinary(binary []byte, contentType string) {
	cb.AddAttachment(base64.StdEncoding.EncodeToString(binary),
		WithFmtType(contentType), WithEncoding("base64"), WithValue("binary"),
	)
}

func (cb *ComponentBase) AddComment(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyComment, s, params...)
}

func (cb *ComponentBase) AddCategory(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyCategories, s, params...)
}

type Attendee struct {
	IANAProperty
}

func (p *Attendee) Email() string {
	if strings.HasPrefix(p.Value, "mailto:") {
		return p.Value[len("mailto:"):]
	}
	return p.Value
}

func (p *Attendee) ParticipationStatus() ParticipationStatus {
	return ParticipationStatus(p.getPropertyFirst(ParameterParticipationStatus))
}

func (p *Attendee) getPropertyFirst(parameter Parameter) string {
	vs := p.getProperty(parameter)
	if len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (p *Attendee) getProperty(parameter Parameter) []string {
	if vs, ok := p.ICalParameters[string(parameter)]; ok {
		return vs
	}
	return nil
}

func (cb *ComponentBase) Attendees() []*Attendee {
	var r []*Attendee
	for i := range cb.Properties {
		switch cb.Properties[i].IANAToken {
		case string(ComponentPropertyAttendee):
			a := &Attendee{
				cb.Properties[i],
			}
			r = append(r, a)
		}
	}
	return r
}

func (cb *ComponentBase) Id() string {
	p := cb.GetProperty(ComponentPropertyUniqueId)
	if p != nil {
		return FromText(p.Value)
	}
	return ""
}

func (cb *ComponentBase) addAlarm() *VAlarm {
	a := &VAlarm{
		ComponentBase: ComponentBase{},
	}
	cb.Components = append(cb.Components, a)
	return a
}

func (cb *ComponentBase) addVAlarm(a *VAlarm) {
	cb.Components = append(cb.Components, a)
}

func (cb *ComponentBase) alarms() []*VAlarm {
	var r []*VAlarm
	for i := range cb.Components {
		switch alarm := cb.Components[i].(type) {
		case *VAlarm:
			r = append(r, alarm)
		}
	}
	return r
}

func (cb *ComponentBase) SetDurationStr(duration string) error {
	_, err := ParseDuration(duration)
	if err != nil {
		return err
	}
	cb.SetProperty(ComponentPropertyDuration, duration)
	return nil
}

type VEvent struct {
	ComponentBase
}

func (event *VEvent) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return event.ComponentBase.serializeThis(w, ComponentVEvent, serialConfig)
}

func (event *VEvent) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := event.serialize(serialConfig)
	return s
}

func (event *VEvent) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := event.ComponentBase.serializeThis(b, ComponentVEvent, serialConfig)
	return b.String(), err
}

func NewEvent(uniqueId string) *VEvent {
	e := &VEvent{
		NewComponent(uniqueId),
	}
	return e
}

func (event *VEvent) SetEndAt(t time.Time, props ...PropertyParameter) {
	layout, props := timestampFormatForTime(t, props)
	event.SetProperty(ComponentPropertyDtEnd, t.Format(layout), props...)
}

func (event *VEvent) SetLastModifiedAt(t time.Time, props ...PropertyParameter) {
	layout, props := timestampFormatForTime(t, props)
	event.SetProperty(ComponentPropertyLastModified, t.Format(layout), props...)
}

// TODO use generics
func (event *VEvent) SetGeo(lat interface{}, lng interface{}, params ...PropertyParameter) {
	event.setGeo(lat, lng, params...)
}

func (event *VEvent) SetPriority(p int, params ...PropertyParameter) {
	event.setPriority(p, params...)
}

func (event *VEvent) SetResources(r string, params ...PropertyParameter) {
	event.setResources(r, params...)
}

func (event *VEvent) AddAlarm() *VAlarm {
	return event.addAlarm()
}

func (event *VEvent) AddVAlarm(a *VAlarm) {
	event.addVAlarm(a)
}

func (event *VEvent) Alarms() []*VAlarm {
	return event.alarms()
}

func (event *VEvent) GetAllDayEndAt() (time.Time, error) {
	return event.getTimeProp(ComponentPropertyDtEnd, true)
}

type TimeTransparency string

const (
	TransparencyOpaque      TimeTransparency = "OPAQUE" // default
	TransparencyTransparent TimeTransparency = "TRANSPARENT"
)

func (event *VEvent) SetTimeTransparency(v TimeTransparency, params ...PropertyParameter) {
	event.SetProperty(ComponentPropertyTransp, string(v), params...)
}

type VTodo struct {
	ComponentBase
}

func (todo *VTodo) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return todo.ComponentBase.serializeThis(w, ComponentVTodo, serialConfig)
}

func (todo *VTodo) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := todo.serialize(serialConfig)
	return s
}

func (todo *VTodo) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := todo.ComponentBase.serializeThis(b, ComponentVTodo, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func NewTodo(uniqueId string) *VTodo {
	e := &VTodo{
		NewComponent(uniqueId),
	}
	return e
}

func (cal *Calendar) AddTodo(id string) *VTodo {
	e := NewTodo(id)
	cal.Components = append(cal.Components, e)
	return e
}

func (cal *Calendar) AddVTodo(e *VTodo) {
	cal.Components = append(cal.Components, e)
}

func (cal *Calendar) Todos() []*VTodo {
	var r []*VTodo
	for i := range cal.Components {
		switch todo := cal.Components[i].(type) {
		case *VTodo:
			r = append(r, todo)
		}
	}
	return r
}

func (todo *VTodo) SetCompletedAt(t time.Time, params ...PropertyParameter) {
	layout, params := timestampFormatForTime(t, params)
	todo.SetProperty(ComponentPropertyCompleted, t.Format(layout), params...)
}

func (todo *VTodo) SetAllDayCompletedAt(t time.Time, params ...PropertyParameter) {
	layout, params := dateFormatForTime(t, params)
	todo.SetProperty(ComponentPropertyCompleted, t.Format(layout), params...)
}

func (todo *VTodo) SetDueAt(t time.Time, params ...PropertyParameter) {
	layout, params := timestampFormatForTime(t, params)
	todo.SetProperty(ComponentPropertyDue, t.Format(layout), params...)
}

func (todo *VTodo) SetAllDayDueAt(t time.Time, params ...PropertyParameter) {
	layout, params := dateFormatForTime(t, params)
	todo.SetProperty(ComponentPropertyDue, t.Format(layout), params...)
}

func (todo *VTodo) SetPercentComplete(p int, params ...PropertyParameter) {
	todo.SetProperty(ComponentPropertyPercentComplete, strconv.Itoa(p), params...)
}

func (todo *VTodo) SetGeo(lat interface{}, lng interface{}, params ...PropertyParameter) {
	todo.setGeo(lat, lng, params...)
}

func (todo *VTodo) SetPriority(p int, params ...PropertyParameter) {
	todo.setPriority(p, params...)
}

func (todo *VTodo) SetResources(r string, params ...PropertyParameter) {
	todo.setResources(r, params...)
}

// SetDuration updates the duration of an event.
// This function will set either the end or start time of an event depending what is already given.
// The duration defines the length of a event relative to start or end time.
//
// Notice: It will not set the DURATION key of the ics - only DTSTART and DTEND will be affected.
func (todo *VTodo) SetDuration(d time.Duration) error {
	t, err := todo.GetStartAt()
	if err == nil {
		todo.SetDueAt(t.Add(d))
		return nil
	} else {
		t, err = todo.GetDueAt()
		if err == nil {
			todo.SetStartAt(t.Add(-d))
			return nil
		}
	}
	return errors.New("start or end not yet defined")
}

func (todo *VTodo) AddAlarm() *VAlarm {
	return todo.addAlarm()
}

func (todo *VTodo) AddVAlarm(a *VAlarm) {
	todo.addVAlarm(a)
}

func (todo *VTodo) Alarms() []*VAlarm {
	return todo.alarms()
}

func (todo *VTodo) GetDueAt() (time.Time, error) {
	return todo.getTimeProp(ComponentPropertyDue, false)
}

func (todo *VTodo) GetAllDayDueAt() (time.Time, error) {
	return todo.getTimeProp(ComponentPropertyDue, true)
}

type VJournal struct {
	ComponentBase
}

func (journal *VJournal) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return journal.ComponentBase.serializeThis(w, ComponentVJournal, serialConfig)
}

func (journal *VJournal) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := journal.serialize(serialConfig)
	return s
}

func (journal *VJournal) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := journal.ComponentBase.serializeThis(b, ComponentVJournal, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func NewJournal(uniqueId string) *VJournal {
	e := &VJournal{
		NewComponent(uniqueId),
	}
	return e
}

func (cal *Calendar) AddJournal(id string) *VJournal {
	e := NewJournal(id)
	cal.Components = append(cal.Components, e)
	return e
}

func (cal *Calendar) AddVJournal(e *VJournal) {
	cal.Components = append(cal.Components, e)
}

func (cal *Calendar) Journals() []*VJournal {
	var r []*VJournal
	for i := range cal.Components {
		switch journal := cal.Components[i].(type) {
		case *VJournal:
			r = append(r, journal)
		}
	}
	return r
}

type VBusy struct {
	ComponentBase
}

func (busy *VBusy) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := busy.serialize(serialConfig)
	return s
}

func (busy *VBusy) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := busy.ComponentBase.serializeThis(b, ComponentVFreeBusy, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (busy *VBusy) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return busy.ComponentBase.serializeThis(w, ComponentVFreeBusy, serialConfig)
}

func NewBusy(uniqueId string) *VBusy {
	e := &VBusy{
		NewComponent(uniqueId),
	}
	return e
}

func (cal *Calendar) AddBusy(id string) *VBusy {
	e := NewBusy(id)
	cal.Components = append(cal.Components, e)
	return e
}

func (cal *Calendar) AddVBusy(e *VBusy) {
	cal.Components = append(cal.Components, e)
}

func (cal *Calendar) Busys() []*VBusy {
	var r []*VBusy
	for i := range cal.Components {
		switch busy := cal.Components[i].(type) {
		case *VBusy:
			r = append(r, busy)
		}
	}
	return r
}

type VTimezone struct {
	ComponentBase
}

func (timezone *VTimezone) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := timezone.serialize(serialConfig)
	return s
}

func (timezone *VTimezone) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := timezone.ComponentBase.serializeThis(b, ComponentVTimezone, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (timezone *VTimezone) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return timezone.ComponentBase.serializeThis(w, ComponentVTimezone, serialConfig)
}

func (timezone *VTimezone) AddStandard() *Standard {
	e := NewStandard()
	timezone.Components = append(timezone.Components, e)
	return e
}

func NewTimezone(tzId string) *VTimezone {
	e := &VTimezone{
		ComponentBase{
			Properties: []IANAProperty{
				{BaseProperty{IANAToken: string(ComponentPropertyTzid), Value: tzId}},
			},
		},
	}
	return e
}

func (cal *Calendar) AddTimezone(id string) *VTimezone {
	e := NewTimezone(id)
	cal.Components = append(cal.Components, e)
	return e
}

func (cal *Calendar) AddVTimezone(e *VTimezone) {
	cal.Components = append(cal.Components, e)
}

func (cal *Calendar) Timezones() []*VTimezone {
	var r []*VTimezone
	for i := range cal.Components {
		switch timezone := cal.Components[i].(type) {
		case *VTimezone:
			r = append(r, timezone)
		}
	}
	return r
}

type VAlarm struct {
	ComponentBase
}

func (c *VAlarm) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := c.serialize(serialConfig)
	return s
}

func (c *VAlarm) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := c.ComponentBase.serializeThis(b, ComponentVAlarm, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (c *VAlarm) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return c.ComponentBase.serializeThis(w, ComponentVAlarm, serialConfig)
}

func NewAlarm(tzId string) *VAlarm {
	// Todo How did this come about?
	e := &VAlarm{}
	return e
}

func (cal *Calendar) AddVAlarm(e *VAlarm) {
	cal.Components = append(cal.Components, e)
}

func (cal *Calendar) Alarms() []*VAlarm {
	var r []*VAlarm
	for i := range cal.Components {
		switch alarm := cal.Components[i].(type) {
		case *VAlarm:
			r = append(r, alarm)
		}
	}
	return r
}

func (c *VAlarm) SetAction(a Action, params ...PropertyParameter) {
	c.SetProperty(ComponentPropertyAction, string(a), params...)
}

func (c *VAlarm) SetTrigger(s string, params ...PropertyParameter) {
	c.SetProperty(ComponentPropertyTrigger, s, params...)
}

type Standard struct {
	ComponentBase
}

func NewStandard() *Standard {
	e := &Standard{
		ComponentBase{},
	}
	return e
}

func (standard *Standard) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := standard.serialize(serialConfig)
	return s
}

func (standard *Standard) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := standard.ComponentBase.serializeThis(b, ComponentStandard, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (standard *Standard) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return standard.ComponentBase.serializeThis(w, ComponentStandard, serialConfig)
}

type Daylight struct {
	ComponentBase
}

func (daylight *Daylight) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := daylight.serialize(serialConfig)
	return s
}

func (daylight *Daylight) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := daylight.ComponentBase.serializeThis(b, ComponentDaylight, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (daylight *Daylight) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return daylight.ComponentBase.serializeThis(w, ComponentDaylight, serialConfig)
}

type GeneralComponent struct {
	ComponentBase
	Token string
}

func (general *GeneralComponent) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := general.serialize(serialConfig)
	return s
}

func (general *GeneralComponent) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &bytes.Buffer{}
	err := general.ComponentBase.serializeThis(b, ComponentType(general.Token), serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (general *GeneralComponent) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return general.ComponentBase.serializeThis(w, ComponentType(general.Token), serialConfig)
}

func GeneralParseComponent(cs *CalendarStream, startLine *BaseProperty) (Component, error) {
	var co Component
	var err error
	switch ComponentType(startLine.Value) {
	case ComponentVCalendar:
		return nil, errors.New("malformed calendar; vcalendar not where expected")
	case ComponentVEvent:
		co, err = ParseVEventWithError(cs, startLine)
	case ComponentVTodo:
		co, err = ParseVTodoWithError(cs, startLine)
	case ComponentVJournal:
		co, err = ParseVJournalWithError(cs, startLine)
	case ComponentVFreeBusy:
		co, err = ParseVBusyWithError(cs, startLine)
	case ComponentVTimezone:
		co, err = ParseVTimezoneWithError(cs, startLine)
	case ComponentVAlarm:
		co, err = ParseVAlarmWithError(cs, startLine)
	case ComponentStandard:
		co, err = ParseStandardWithError(cs, startLine)
	case ComponentDaylight:
		co, err = ParseDaylightWithError(cs, startLine)
	default:
		co, err = ParseGeneralComponentWithError(cs, startLine)
	}
	return co, err
}

func ParseVEvent(cs *CalendarStream, startLine *BaseProperty) *VEvent {
	ev, _ := ParseVEventWithError(cs, startLine)
	return ev
}

func ParseVEventWithError(cs *CalendarStream, startLine *BaseProperty) (*VEvent, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}
	rr := &VEvent{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseVTodo(cs *CalendarStream, startLine *BaseProperty) *VTodo {
	c, _ := ParseVTodoWithError(cs, startLine)
	return c
}

func ParseVTodoWithError(cs *CalendarStream, startLine *BaseProperty) (*VTodo, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &VTodo{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseVJournal(cs *CalendarStream, startLine *BaseProperty) *VJournal {
	c, _ := ParseVJournalWithError(cs, startLine)
	return c
}

func ParseVJournalWithError(cs *CalendarStream, startLine *BaseProperty) (*VJournal, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &VJournal{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseVBusy(cs *CalendarStream, startLine *BaseProperty) *VBusy {
	c, _ := ParseVBusyWithError(cs, startLine)
	return c
}

func ParseVBusyWithError(cs *CalendarStream, startLine *BaseProperty) (*VBusy, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &VBusy{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseVTimezone(cs *CalendarStream, startLine *BaseProperty) *VTimezone {
	c, _ := ParseVTimezoneWithError(cs, startLine)
	return c
}

func ParseVTimezoneWithError(cs *CalendarStream, startLine *BaseProperty) (*VTimezone, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &VTimezone{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseVAlarm(cs *CalendarStream, startLine *BaseProperty) *VAlarm {
	c, _ := ParseVAlarmWithError(cs, startLine)
	return c
}

func ParseVAlarmWithError(cs *CalendarStream, startLine *BaseProperty) (*VAlarm, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &VAlarm{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseStandard(cs *CalendarStream, startLine *BaseProperty) *Standard {
	c, _ := ParseStandardWithError(cs, startLine)
	return c
}

func ParseStandardWithError(cs *CalendarStream, startLine *BaseProperty) (*Standard, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &Standard{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseDaylight(cs *CalendarStream, startLine *BaseProperty) *Daylight {
	c, _ := ParseDaylightWithError(cs, startLine)
	return c
}

func ParseDaylightWithError(cs *CalendarStream, startLine *BaseProperty) (*Daylight, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &Daylight{
		ComponentBase: r,
	}
	return rr, nil
}

func ParseGeneralComponent(cs *CalendarStream, startLine *BaseProperty) *GeneralComponent {
	c, _ := ParseGeneralComponentWithError(cs, startLine)
	return c
}

func ParseGeneralComponentWithError(cs *CalendarStream, startLine *BaseProperty) (*GeneralComponent, error) {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil, err
	}
	rr := &GeneralComponent{
		ComponentBase: r,
		Token:         startLine.Value,
	}
	return rr, nil
}

func ParseComponent(cs *CalendarStream, startLine *BaseProperty) (ComponentBase, error) {
	cb := ComponentBase{}
	cont := true
	for ln := 0; cont; ln++ {
		l, err := cs.ReadLine()
		if err != nil {
			switch err {
			case io.EOF:
				cont = false
			default:
				return cb, err
			}
		}
		if l == nil || len(*l) == 0 {
			continue
		}
		line, err := ParseProperty(*l)
		if err != nil {
			return cb, fmt.Errorf("parsing component property %d: %w", ln, err)
		}
		if line == nil {
			return cb, errors.New("parsing component line")
		}
		switch line.IANAToken {
		case "END":
			switch line.Value {
			case startLine.Value:
				return cb, nil
			default:
				return cb, errors.New("unbalanced end")
			}
		case "BEGIN":
			co, err := GeneralParseComponent(cs, line)
			if err != nil {
				return cb, err
			}
			if co != nil {
				cb.Components = append(cb.Components, co)
			}
		default: // TODO put in all the supported types for type switching etc.
			cb.Properties = append(cb.Properties, IANAProperty{*line})
		}
	}
	return cb, errors.New("ran out of lines")
}
