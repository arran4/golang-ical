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

type Component interface {
	UnknownPropertiesIANAProperties() []IANAProperty
	SubComponents() []Component
	serialize(b io.Writer)
}

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
		c.serialize(writer)
	}
	_, _ = fmt.Fprint(writer, "END:"+componentType, "\r\n")
}

func NewComponent(uniqueId string) ComponentBase {
	return ComponentBase{
		Properties: []IANAProperty{
			{BaseProperty{IANAToken: ToText(string(ComponentPropertyUniqueId)), Value: uniqueId}},
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

func (cb *ComponentBase) SetProperty(property ComponentProperty, value string, props ...PropertyParameter) {
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken == string(property) {
			cb.Properties[i].Value = value
			cb.Properties[i].ICalParameters = map[string][]string{}
			for _, p := range props {
				k, v := p.KeyValue()
				cb.Properties[i].ICalParameters[k] = v
			}
			return
		}
	}
	cb.AddProperty(property, value, props...)
}

func (cb *ComponentBase) AddProperty(property ComponentProperty, value string, props ...PropertyParameter) {
	r := IANAProperty{
		BaseProperty{
			IANAToken:      string(property),
			Value:          value,
			ICalParameters: map[string][]string{},
		},
	}
	for _, p := range props {
		k, v := p.KeyValue()
		r.ICalParameters[k] = v
	}
	cb.Properties = append(cb.Properties, r)
}

const (
	icalTimestampFormatUtc   = "20060102T150405Z"
	icalTimestampFormatLocal = "20060102T150405"
	icalDateFormatUtc        = "20060102Z"
	icalDateFormatLocal      = "20060102"
)

var (
	timeStampVariations = regexp.MustCompile("^([0-9]{8})?([TZ])?([0-9]{6})?(Z)?$")
)

func (cb *ComponentBase) SetCreatedTime(t time.Time, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyCreated, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (cb *ComponentBase) SetDtStampTime(t time.Time, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDtstamp, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (cb *ComponentBase) SetModifiedAt(t time.Time, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyLastModified, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (cb *ComponentBase) SetSequence(seq int, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertySequence, strconv.Itoa(seq), props...)
}

func (cb *ComponentBase) SetStartAt(t time.Time, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDtStart, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (cb *ComponentBase) SetAllDayStartAt(t time.Time, props ...PropertyParameter) {
	props = append(props, WithValue(string(ValueDataTypeDate)))
	cb.SetProperty(ComponentPropertyDtStart, t.Format(icalDateFormatLocal), props...)
}

func (cb *ComponentBase) SetAllDayEndAt(t time.Time, props ...PropertyParameter) {
	props = append(props, WithValue(string(ValueDataTypeDate)))
	cb.SetProperty(ComponentPropertyDtEnd, t.Format(icalDateFormatLocal), props...)
}

func (cb *ComponentBase) getTimeProp(componentProperty ComponentProperty, expectAllDay bool) (time.Time, error) {
	timeProp := cb.GetProperty(componentProperty)
	if timeProp == nil {
		return time.Time{}, errors.New("property not found")
	}

	timeVal := timeProp.BaseProperty.Value
	matched := timeStampVariations.FindStringSubmatch(timeVal)
	if matched == nil {
		return time.Time{}, fmt.Errorf("time value not matched, got '%s'", timeVal)
	}
	tOrZGrp := matched[2]
	zGrp := matched[4]
	grp1len := len(matched[1])
	grp3len := len(matched[3])

	tzId, tzIdOk := timeProp.ICalParameters["TZID"]
	var propLoc *time.Location
	if tzIdOk {
		if len(tzId) != 1 {
			return time.Time{}, errors.New("expected only one TZID")
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

		return time.Time{}, fmt.Errorf("time value matched but unsupported all-day timestamp, got '%s'", timeVal)
	}

	if grp1len > 0 && grp3len > 0 && tOrZGrp == "T" && zGrp == "Z" {
		return time.ParseInLocation(icalTimestampFormatUtc, timeVal, time.UTC)
	} else if grp1len > 0 && grp3len > 0 && tOrZGrp == "T" && zGrp == "" {
		if propLoc == nil {
			return time.ParseInLocation(icalTimestampFormatLocal, timeVal, time.Local)
		} else {
			return time.ParseInLocation(icalTimestampFormatLocal, timeVal, propLoc)
		}
	} else if grp1len > 0 && grp3len == 0 && tOrZGrp == "Z" && zGrp == "" {
		return time.ParseInLocation(icalDateFormatUtc, dateStr+"Z", time.UTC)
	} else if grp1len > 0 && grp3len == 0 && tOrZGrp == "" && zGrp == "" {
		if propLoc == nil {
			return time.ParseInLocation(icalDateFormatLocal, dateStr, time.Local)
		} else {
			return time.ParseInLocation(icalDateFormatLocal, dateStr, propLoc)
		}
	}

	return time.Time{}, fmt.Errorf("time value matched but not supported, got '%s'", timeVal)
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

func (cb *ComponentBase) SetSummary(s string, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertySummary, s, props...)
}

func (cb *ComponentBase) SetStatus(s ObjectStatus, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyStatus, string(s), props...)
}

func (cb *ComponentBase) SetDescription(s string, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDescription, s, props...)
}

func (cb *ComponentBase) SetLocation(s string, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyLocation, s, props...)
}

func (cb *ComponentBase) setGeo(lat interface{}, lng interface{}, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyGeo, fmt.Sprintf("%v;%v", lat, lng), props...)
}

func (cb *ComponentBase) SetURL(s string, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyUrl, s, props...)
}

func (cb *ComponentBase) SetOrganizer(s string, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyOrganizer, s, props...)
}

func (cb *ComponentBase) SetColor(s string, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyColor, s, props...)
}

func (cb *ComponentBase) SetClass(c Classification, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyClass, string(c), props...)
}

func (cb *ComponentBase) setPriority(p int, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyPriority, strconv.Itoa(p), props...)
}

func (cb *ComponentBase) setResources(r string, props ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyResources, r, props...)
}

func (cb *ComponentBase) AddAttendee(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyAttendee, "mailto:"+s, props...)
}

func (cb *ComponentBase) AddExdate(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyExdate, s, props...)
}

func (cb *ComponentBase) AddExrule(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyExrule, s, props...)
}

func (cb *ComponentBase) AddRdate(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyRdate, s, props...)
}

func (cb *ComponentBase) AddRrule(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyRrule, s, props...)
}

func (cb *ComponentBase) AddAttachment(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyAttach, s, props...)
}

func (cb *ComponentBase) AddAttachmentURL(uri string, contentType string) {
	cb.AddAttachment(uri, WithFmtType(contentType))
}

func (cb *ComponentBase) AddAttachmentBinary(binary []byte, contentType string) {
	cb.AddAttachment(base64.StdEncoding.EncodeToString(binary),
		WithFmtType(contentType), WithEncoding("base64"), WithValue("binary"),
	)
}

func (cb *ComponentBase) AddComment(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyComment, s, props...)
}

func (cb *ComponentBase) AddCategory(s string, props ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyCategories, s, props...)
}

type Attendee struct {
	IANAProperty
}

func (attendee *Attendee) Email() string {
	if strings.HasPrefix(attendee.Value, "mailto:") {
		return attendee.Value[len("mailto:"):]
	}
	return attendee.Value
}

func (attendee *Attendee) ParticipationStatus() ParticipationStatus {
	return ParticipationStatus(attendee.getPropertyFirst(ParameterParticipationStatus))
}

func (attendee *Attendee) getPropertyFirst(parameter Parameter) string {
	vs := attendee.getProperty(parameter)
	if len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (attendee *Attendee) getProperty(parameter Parameter) []string {
	if vs, ok := attendee.ICalParameters[string(parameter)]; ok {
		return vs
	}
	return nil
}

func (cb *ComponentBase) Attendees() (r []*Attendee) {
	r = []*Attendee{}
	for i := range cb.Properties {
		switch cb.Properties[i].IANAToken {
		case string(ComponentPropertyAttendee):
			a := &Attendee{
				cb.Properties[i],
			}
			r = append(r, a)
		}
	}
	return
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

func (cb *ComponentBase) alarms() (r []*VAlarm) {
	r = []*VAlarm{}
	for i := range cb.Components {
		switch alarm := cb.Components[i].(type) {
		case *VAlarm:
			r = append(r, alarm)
		}
	}
	return
}

type VEvent struct {
	ComponentBase
}

func (c *VEvent) serialize(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VEVENT")
}

func (c *VEvent) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VEVENT")
	return b.String()
}

func NewEvent(uniqueId string) *VEvent {
	e := &VEvent{
		NewComponent(uniqueId),
	}
	return e
}

func (calendar *Calendar) AddEvent(id string) *VEvent {
	e := NewEvent(id)
	calendar.Components = append(calendar.Components, e)
	return e
}

func (calendar *Calendar) AddVEvent(e *VEvent) {
	calendar.Components = append(calendar.Components, e)
}

func (calendar *Calendar) Events() (r []*VEvent) {
	r = []*VEvent{}
	for i := range calendar.Components {
		switch event := calendar.Components[i].(type) {
		case *VEvent:
			r = append(r, event)
		}
	}
	return
}

func (event *VEvent) SetEndAt(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyDtEnd, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetLastModifiedAt(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyLastModified, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetGeo(lat interface{}, lng interface{}, props ...PropertyParameter) {
	event.setGeo(lat, lng, props...)
}

func (event *VEvent) SetPriority(p int, props ...PropertyParameter) {
	event.setPriority(p, props...)
}

func (event *VEvent) SetResources(r string, props ...PropertyParameter) {
	event.setResources(r, props...)
}

// SetDuration updates the duration of an event.
// This function will set either the end or start time of an event depending what is already given.
// The duration defines the length of a event relative to start or end time.
//
// Notice: It will not set the DURATION key of the ics - only DTSTART and DTEND will be affected.
func (event *VEvent) SetDuration(d time.Duration) error {
	t, err := event.GetStartAt()
	if err == nil {
		event.SetEndAt(t.Add(d))
		return nil
	} else {
		t, err = event.GetEndAt()
		if err == nil {
			event.SetStartAt(t.Add(-d))
			return nil
		}
	}
	return errors.New("start or end not yet defined")
}

func (event *VEvent) AddAlarm() *VAlarm {
	return event.addAlarm()
}

func (event *VEvent) AddVAlarm(a *VAlarm) {
	event.addVAlarm(a)
}

func (event *VEvent) Alarms() (r []*VAlarm) {
	return event.alarms()
}

func (event *VEvent) GetEndAt() (time.Time, error) {
	return event.getTimeProp(ComponentPropertyDtEnd, false)
}

func (event *VEvent) GetAllDayEndAt() (time.Time, error) {
	return event.getTimeProp(ComponentPropertyDtEnd, true)
}

type TimeTransparency string

const (
	TransparencyOpaque      TimeTransparency = "OPAQUE" // default
	TransparencyTransparent TimeTransparency = "TRANSPARENT"
)

func (event *VEvent) SetTimeTransparency(v TimeTransparency, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyTransp, string(v), props...)
}

type VTodo struct {
	ComponentBase
}

func (todo *VTodo) serialize(w io.Writer) {
	todo.ComponentBase.serializeThis(w, "VTODO")
}

func (todo *VTodo) Serialize() string {
	b := &bytes.Buffer{}
	todo.ComponentBase.serializeThis(b, "VTODO")
	return b.String()
}

func NewTodo(uniqueId string) *VTodo {
	e := &VTodo{
		NewComponent(uniqueId),
	}
	return e
}

func (calendar *Calendar) AddTodo(id string) *VTodo {
	e := NewTodo(id)
	calendar.Components = append(calendar.Components, e)
	return e
}

func (calendar *Calendar) AddVTodo(e *VTodo) {
	calendar.Components = append(calendar.Components, e)
}

func (calendar *Calendar) Todos() (r []*VTodo) {
	r = []*VTodo{}
	for i := range calendar.Components {
		switch todo := calendar.Components[i].(type) {
		case *VTodo:
			r = append(r, todo)
		}
	}
	return
}

func (todo *VTodo) SetCompletedAt(t time.Time, props ...PropertyParameter) {
	todo.SetProperty(ComponentPropertyCompleted, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (todo *VTodo) SetAllDayCompletedAt(t time.Time, props ...PropertyParameter) {
	props = append(props, WithValue(string(ValueDataTypeDate)))
	todo.SetProperty(ComponentPropertyCompleted, t.Format(icalDateFormatLocal), props...)
}

func (todo *VTodo) SetDueAt(t time.Time, props ...PropertyParameter) {
	todo.SetProperty(ComponentPropertyDue, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (todo *VTodo) SetAllDayDueAt(t time.Time, props ...PropertyParameter) {
	props = append(props, WithValue(string(ValueDataTypeDate)))
	todo.SetProperty(ComponentPropertyDue, t.Format(icalDateFormatLocal), props...)
}

func (todo *VTodo) SetPercentComplete(p int, props ...PropertyParameter) {
	todo.SetProperty(ComponentPropertyPercentComplete, strconv.Itoa(p), props...)
}

func (todo *VTodo) SetGeo(lat interface{}, lng interface{}, props ...PropertyParameter) {
	todo.setGeo(lat, lng, props...)
}

func (todo *VTodo) SetPriority(p int, props ...PropertyParameter) {
	todo.setPriority(p, props...)
}

func (todo *VTodo) SetResources(r string, props ...PropertyParameter) {
	todo.setResources(r, props...)
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

func (todo *VTodo) Alarms() (r []*VAlarm) {
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

func (c *VJournal) serialize(w io.Writer) {
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

func (calendar *Calendar) AddJournal(id string) *VJournal {
	e := NewJournal(id)
	calendar.Components = append(calendar.Components, e)
	return e
}

func (calendar *Calendar) AddVJournal(e *VJournal) {
	calendar.Components = append(calendar.Components, e)
}

func (calendar *Calendar) Journals() (r []*VJournal) {
	r = []*VJournal{}
	for i := range calendar.Components {
		switch journal := calendar.Components[i].(type) {
		case *VJournal:
			r = append(r, journal)
		}
	}
	return
}

type VBusy struct {
	ComponentBase
}

func (c *VBusy) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VBUSY")
	return b.String()
}

func (c *VBusy) serialize(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VBUSY")
}

func NewBusy(uniqueId string) *VBusy {
	e := &VBusy{
		NewComponent(uniqueId),
	}
	return e
}

func (calendar *Calendar) AddBusy(id string) *VBusy {
	e := NewBusy(id)
	calendar.Components = append(calendar.Components, e)
	return e
}

func (calendar *Calendar) AddVBusy(e *VBusy) {
	calendar.Components = append(calendar.Components, e)
}

func (calendar *Calendar) Busys() (r []*VBusy) {
	r = []*VBusy{}
	for i := range calendar.Components {
		switch busy := calendar.Components[i].(type) {
		case *VBusy:
			r = append(r, busy)
		}
	}
	return
}

type VTimezone struct {
	ComponentBase
}

func (c *VTimezone) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VTIMEZONE")
	return b.String()
}

func (c *VTimezone) serialize(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VTIMEZONE")
}

func NewTimezone(tzId string) *VTimezone {
	e := &VTimezone{
		ComponentBase{
			Properties: []IANAProperty{
				{BaseProperty{IANAToken: ToText(string(ComponentPropertyTzid)), Value: tzId}},
			},
		},
	}
	return e
}

func (calendar *Calendar) AddTimezone(id string) *VTimezone {
	e := NewTimezone(id)
	calendar.Components = append(calendar.Components, e)
	return e
}

func (calendar *Calendar) AddVTimezone(e *VTimezone) {
	calendar.Components = append(calendar.Components, e)
}

func (calendar *Calendar) Timezones() (r []*VTimezone) {
	r = []*VTimezone{}
	for i := range calendar.Components {
		switch timezone := calendar.Components[i].(type) {
		case *VTimezone:
			r = append(r, timezone)
		}
	}
	return
}

type VAlarm struct {
	ComponentBase
}

func (c *VAlarm) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VALARM")
	return b.String()
}

func (c *VAlarm) serialize(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VALARM")
}

func NewAlarm(tzId string) *VAlarm {
	e := &VAlarm{}
	return e
}

func (calendar *Calendar) AddVAlarm(e *VAlarm) {
	calendar.Components = append(calendar.Components, e)
}

func (calendar *Calendar) Alarms() (r []*VAlarm) {
	r = []*VAlarm{}
	for i := range calendar.Components {
		switch alarm := calendar.Components[i].(type) {
		case *VAlarm:
			r = append(r, alarm)
		}
	}
	return
}

func (alarm *VAlarm) SetAction(a Action, props ...PropertyParameter) {
	alarm.SetProperty(ComponentPropertyAction, string(a), props...)
}

func (alarm *VAlarm) SetTrigger(s string, props ...PropertyParameter) {
	alarm.SetProperty(ComponentPropertyTrigger, s, props...)
}

type Standard struct {
	ComponentBase
}

func (c *Standard) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "STANDARD")
	return b.String()
}

func (c *Standard) serialize(w io.Writer) {
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

func (c *Daylight) serialize(w io.Writer) {
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

func (c *GeneralComponent) serialize(w io.Writer) {
	c.ComponentBase.serializeThis(w, c.Token)
}

func GeneralParseComponent(cs *CalendarStream, startLine *BaseProperty) (Component, error) {
	var co Component
	switch startLine.Value {
	case "VCALENDAR":
		return nil, errors.New("malformed calendar; vcalendar not where expected")
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
