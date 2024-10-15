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
	SerializeTo(b io.Writer)
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

func (cb ComponentBase) serializeThis(writer io.Writer, componentType string) {
	_, _ = fmt.Fprint(writer, "BEGIN:"+componentType, "\r\n")
	for _, p := range cb.Properties {
		p.serialize(writer)
	}
	for _, c := range cb.Components {
		c.SerializeTo(writer)
	}
	_, _ = fmt.Fprint(writer, "END:"+componentType, "\r\n")
}

func NewComponent(uniqueId string) ComponentBase {
	return ComponentBase{
		Properties: []IANAProperty{
			{BaseProperty{IANAToken: string(ComponentPropertyUniqueId), Value: uniqueId}},
		},
	}
}

func (cb *ComponentBase) GetProperty(componentProperty ComponentProperty) *IANAProperty {
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken == string(componentProperty) {
			return &cb.Properties[i]
		}
	}
	return nil
}

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

// RemoveProperty removes from the component all properties that has
// the name passed in removeProp.
func (cb *ComponentBase) RemoveProperty(removeProp ComponentProperty) {
	var keptProperties []IANAProperty
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken != string(removeProp) {
			keptProperties = append(keptProperties, cb.Properties[i])
		}
	}
	cb.Properties = keptProperties
}

const (
	icalTimestampFormatUtc   = "20060102T150405Z"
	icalTimestampFormatLocal = "20060102T150405"
	icalDateFormatUtc        = "20060102Z"
	icalDateFormatLocal      = "20060102"
)

var timeStampVariations = regexp.MustCompile("^([0-9]{8})?([TZ])?([0-9]{6})?(Z)?$")

func (cb *ComponentBase) SetCreatedTime(t time.Time, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyCreated, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (cb *ComponentBase) SetDtStampTime(t time.Time, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDtstamp, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (cb *ComponentBase) SetModifiedAt(t time.Time, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyLastModified, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (cb *ComponentBase) SetSequence(seq int, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertySequence, strconv.Itoa(seq), params...)
}

func (cb *ComponentBase) SetStartAt(t time.Time, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDtStart, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (cb *ComponentBase) SetAllDayStartAt(t time.Time, params ...PropertyParameter) {
	cb.SetProperty(
		ComponentPropertyDtStart,
		t.Format(icalDateFormatLocal),
		append(params, WithValue(string(ValueDataTypeDate)))...,
	)
}

func (cb *ComponentBase) SetEndAt(t time.Time, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDtEnd, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (cb *ComponentBase) SetAllDayEndAt(t time.Time, params ...PropertyParameter) {
	cb.SetProperty(
		ComponentPropertyDtEnd,
		t.Format(icalDateFormatLocal),
		append(params, WithValue(string(ValueDataTypeDate)))...,
	)
}

// SetDuration updates the duration of an event.
// This function will set either the end or start time of an event depending what is already given.
// The duration defines the length of a event relative to start or end time.
//
// Notice: It will not set the DURATION key of the ics - only DTSTART and DTEND will be affected.
func (cb *ComponentBase) SetDuration(d time.Duration) error {
	startProp := cb.GetProperty(ComponentPropertyDtStart)
	if startProp != nil {
		t, err := cb.GetStartAt()
		if err == nil {
			v, _ := startProp.parameterValue(ParameterValue)
			if v == string(ValueDataTypeDate) {
				cb.SetAllDayEndAt(t.Add(d))
			} else {
				cb.SetEndAt(t.Add(d))
			}
			return nil
		}
	}
	endProp := cb.GetProperty(ComponentPropertyDtEnd)
	if endProp != nil {
		t, err := cb.GetEndAt()
		if err == nil {
			v, _ := endProp.parameterValue(ParameterValue)
			if v == string(ValueDataTypeDate) {
				cb.SetAllDayStartAt(t.Add(-d))
			} else {
				cb.SetStartAt(t.Add(-d))
			}
			return nil
		}
	}
	return errors.New("start or end not yet defined")
}

func (cb *ComponentBase) GetEndAt() (time.Time, error) {
	return cb.getTimeProp(ComponentPropertyDtEnd, false)
}

func (cb *ComponentBase) getTimeProp(componentProperty ComponentProperty, expectAllDay bool) (time.Time, error) {
	timeProp := cb.GetProperty(componentProperty)
	if timeProp == nil {
		return time.Time{}, fmt.Errorf("%w: %s", ErrPropertyNotFound, componentProperty)
	}

	timeVal := timeProp.BaseProperty.Value
	matched := timeStampVariations.FindStringSubmatch(timeVal)
	if matched == nil {
		return time.Time{}, fmt.Errorf("%w, got '%s'", ErrTimeValueNotMatched, timeVal)
	}
	tOrZGrp := matched[2]
	zGrp := matched[4]
	grp1len := len(matched[1])
	grp3len := len(matched[3])

	tzId, tzIdOk := timeProp.ICalParameters["TZID"]
	var propLoc *time.Location
	if tzIdOk {
		if len(tzId) != 1 {
			return time.Time{}, ErrExpectedOneTZID
		}
		var tzErr error
		propLoc, tzErr = time.LoadLocation(tzId[0])
		if tzErr != nil {
			return time.Time{}, tzErr
		}
	}
	dateStr := matched[1]

	if expectAllDay {
		if grp1len > 0 {
			if tOrZGrp == "Z" || zGrp == "Z" {
				return time.ParseInLocation(icalDateFormatUtc, dateStr+"Z", time.UTC)
			} else {
				if propLoc == nil {
					return time.ParseInLocation(icalDateFormatLocal, dateStr, time.Local)
				} else {
					return time.ParseInLocation(icalDateFormatLocal, dateStr, propLoc)
				}
			}
		}

		return time.Time{}, fmt.Errorf("%w, got '%s'", ErrTimeValueMatchedButUnsupportedAllDayTimeStamp, timeVal)
	}

	switch {
	case grp1len > 0 && grp3len > 0 && tOrZGrp == "T" && zGrp == "Z":
		return time.ParseInLocation(icalTimestampFormatUtc, timeVal, time.UTC)
	case grp1len > 0 && grp3len > 0 && tOrZGrp == "T" && zGrp == "":
		if propLoc == nil {
			return time.ParseInLocation(icalTimestampFormatLocal, timeVal, time.Local)
		} else {
			return time.ParseInLocation(icalTimestampFormatLocal, timeVal, propLoc)
		}
	case grp1len > 0 && grp3len == 0 && tOrZGrp == "Z" && zGrp == "":
		return time.ParseInLocation(icalDateFormatUtc, dateStr+"Z", time.UTC)
	case grp1len > 0 && grp3len == 0 && tOrZGrp == "" && zGrp == "":
		if propLoc == nil {
			return time.ParseInLocation(icalDateFormatLocal, dateStr, time.Local)
		} else {
			return time.ParseInLocation(icalDateFormatLocal, dateStr, propLoc)
		}
	}

	return time.Time{}, fmt.Errorf("%w, got '%s'", ErrTimeValueMatchedButNotSupported, timeVal)
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

type VEvent struct {
	ComponentBase
}

func (event *VEvent) SerializeTo(w io.Writer) {
	event.ComponentBase.serializeThis(w, "VEVENT")
}

func (event *VEvent) Serialize() string {
	b := &bytes.Buffer{}
	event.ComponentBase.serializeThis(b, "VEVENT")
	return b.String()
}

func NewEvent(uniqueId string) *VEvent {
	e := &VEvent{
		NewComponent(uniqueId),
	}
	return e
}

func (event *VEvent) SetEndAt(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyDtEnd, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetLastModifiedAt(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyLastModified, t.UTC().Format(icalTimestampFormatUtc), props...)
}

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

func (c *VTodo) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VTODO")
}

func (c *VTodo) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VTODO")
	return b.String()
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

func (c *VTodo) SetCompletedAt(t time.Time, params ...PropertyParameter) {
	c.SetProperty(ComponentPropertyCompleted, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (c *VTodo) SetAllDayCompletedAt(t time.Time, params ...PropertyParameter) {
	params = append(params, WithValue(string(ValueDataTypeDate)))
	c.SetProperty(ComponentPropertyCompleted, t.Format(icalDateFormatLocal), params...)
}

func (c *VTodo) SetDueAt(t time.Time, params ...PropertyParameter) {
	c.SetProperty(ComponentPropertyDue, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (c *VTodo) SetAllDayDueAt(t time.Time, params ...PropertyParameter) {
	params = append(params, WithValue(string(ValueDataTypeDate)))
	c.SetProperty(ComponentPropertyDue, t.Format(icalDateFormatLocal), params...)
}

func (c *VTodo) SetPercentComplete(p int, params ...PropertyParameter) {
	c.SetProperty(ComponentPropertyPercentComplete, strconv.Itoa(p), params...)
}

func (c *VTodo) SetGeo(lat interface{}, lng interface{}, params ...PropertyParameter) {
	c.setGeo(lat, lng, params...)
}

func (c *VTodo) SetPriority(p int, params ...PropertyParameter) {
	c.setPriority(p, params...)
}

func (c *VTodo) SetResources(r string, params ...PropertyParameter) {
	c.setResources(r, params...)
}

// SetDuration updates the duration of an event.
// This function will set either the end or start time of an event depending what is already given.
// The duration defines the length of a event relative to start or end time.
//
// Notice: It will not set the DURATION key of the ics - only DTSTART and DTEND will be affected.
func (c *VTodo) SetDuration(d time.Duration) error {
	t, err := c.GetStartAt()
	if err == nil {
		c.SetDueAt(t.Add(d))
		return nil
	} else {
		t, err = c.GetDueAt()
		if err == nil {
			c.SetStartAt(t.Add(-d))
			return nil
		}
	}
	return errors.New("start or end not yet defined")
}

func (c *VTodo) AddAlarm() *VAlarm {
	return c.addAlarm()
}

func (c *VTodo) AddVAlarm(a *VAlarm) {
	c.addVAlarm(a)
}

func (c *VTodo) Alarms() []*VAlarm {
	return c.alarms()
}

func (cb *ComponentBase) GetDueAt() (time.Time, error) {
	return cb.getTimeProp(ComponentPropertyDue, false)
}

func (c *VEvent) GetAllDayDueAt() (time.Time, error) {
	return c.getTimeProp(ComponentPropertyDue, true)
}

type VJournal struct {
	ComponentBase
}

func (c *VJournal) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VJOURNAL")
}

func (c *VJournal) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VJOURNAL")
	return b.String()
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

func (c *VBusy) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VFREEBUSY")
	return b.String()
}

func (c *VBusy) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VFREEBUSY")
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

func (c *VTimezone) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VTIMEZONE")
	return b.String()
}

func (c *VTimezone) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VTIMEZONE")
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

func (c *VAlarm) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VALARM")
	return b.String()
}

func (c *VAlarm) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VALARM")
}

func NewAlarm(tzId string) *VAlarm {
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

func (c *Standard) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "STANDARD")
	return b.String()
}

func (c *Standard) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, "STANDARD")
}

type Daylight struct {
	ComponentBase
}

func (c *Daylight) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "DAYLIGHT")
	return b.String()
}

func (c *Daylight) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, "DAYLIGHT")
}

type GeneralComponent struct {
	ComponentBase
	Token string
}

func (c *GeneralComponent) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, c.Token)
	return b.String()
}

func (c *GeneralComponent) SerializeTo(w io.Writer) {
	c.ComponentBase.serializeThis(w, c.Token)
}

func GeneralParseComponent(cs *CalendarStream, startLine *BaseProperty) (Component, error) {
	var co Component
	switch startLine.Value {
	case "VCALENDAR":
		return nil, ErrMalformedCalendarVCalendarNotWhereExpected
	case "VEVENT":
		co = ParseVEvent(cs, startLine)
	case "VTODO":
		co = ParseVTodo(cs, startLine)
	case "VJOURNAL":
		co = ParseVJournal(cs, startLine)
	case "VFREEBUSY":
		co = ParseVBusy(cs, startLine)
	case "VTIMEZONE":
		co = ParseVTimezone(cs, startLine)
	case "VALARM":
		co = ParseVAlarm(cs, startLine)
	case "STANDARD":
		co = ParseStandard(cs, startLine)
	case "DAYLIGHT":
		co = ParseDaylight(cs, startLine)
	default:
		co = ParseGeneralComponent(cs, startLine)
	}
	return co, nil
}

func ParseVEvent(cs *CalendarStream, startLine *BaseProperty) *VEvent {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &VEvent{
		ComponentBase: r,
	}
	return rr
}

func ParseVTodo(cs *CalendarStream, startLine *BaseProperty) *VTodo {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &VTodo{
		ComponentBase: r,
	}
	return rr
}

func ParseVJournal(cs *CalendarStream, startLine *BaseProperty) *VJournal {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &VJournal{
		ComponentBase: r,
	}
	return rr
}

func ParseVBusy(cs *CalendarStream, startLine *BaseProperty) *VBusy {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &VBusy{
		ComponentBase: r,
	}
	return rr
}

func ParseVTimezone(cs *CalendarStream, startLine *BaseProperty) *VTimezone {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &VTimezone{
		ComponentBase: r,
	}
	return rr
}

func ParseVAlarm(cs *CalendarStream, startLine *BaseProperty) *VAlarm {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &VAlarm{
		ComponentBase: r,
	}
	return rr
}

func ParseStandard(cs *CalendarStream, startLine *BaseProperty) *Standard {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &Standard{
		ComponentBase: r,
	}
	return rr
}

func ParseDaylight(cs *CalendarStream, startLine *BaseProperty) *Daylight {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &Daylight{
		ComponentBase: r,
	}
	return rr
}

func ParseGeneralComponent(cs *CalendarStream, startLine *BaseProperty) *GeneralComponent {
	r, err := ParseComponent(cs, startLine)
	if err != nil {
		return nil
	}
	rr := &GeneralComponent{
		ComponentBase: r,
		Token:         startLine.Value,
	}
	return rr
}

func ParseComponent(cs *CalendarStream, startLine *BaseProperty) (ComponentBase, error) {
	cb := ComponentBase{}
	cont := true
	for ln := 0; cont; ln++ {
		l, err := cs.ReadLine()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
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
			return cb, fmt.Errorf("%w %d: %w", ErrParsingComponentProperty, ln, err)
		}
		if line == nil {
			return cb, ErrParsingComponentLine
		}
		switch line.IANAToken {
		case "END":
			switch line.Value {
			case startLine.Value:
				return cb, nil
			default:
				return cb, ErrUnbalancedEnd
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
	return cb, ErrOutOfLines
}
