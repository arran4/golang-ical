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
func (base ComponentBase) serializeThis(writer io.Writer, componentType string) {
	fmt.Fprint(writer, "BEGIN:"+componentType, "\r\n")
	for _, p := range base.Properties {
		p.serialize(writer)
	}
	for _, c := range base.Components {
		c.serialize(writer)
	}
	fmt.Fprint(writer, "END:"+componentType, "\r\n")
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

const (
	icalTimestampFormatUtc   = "20060102T150405Z"
	icalTimestampFormatLocal = "20060102T150405"
	icalDateFormatUtc        = "20060102Z"
	icalDateFormatLocal      = "20060102"
)

var (
	timeStampVariations = regexp.MustCompile("^([0-9]{8})?([TZ])?([0-9]{6})?(Z)?$")
)

func (event *VEvent) SetCreatedTime(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyCreated, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetDtStampTime(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyDtstamp, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetModifiedAt(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyLastModified, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetSequence(seq int, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertySequence, strconv.Itoa(seq), props...)
}

func (event *VEvent) SetStartAt(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyDtStart, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetAllDayStartAt(t time.Time, props ...PropertyParameter) {
	props = append(props, WithValue(string(ValueDataTypeDate)))
	event.SetProperty(ComponentPropertyDtStart, t.Format(icalDateFormatLocal), props...)
}

func (event *VEvent) SetEndAt(t time.Time, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyDtEnd, t.UTC().Format(icalTimestampFormatUtc), props...)
}

func (event *VEvent) SetAllDayEndAt(t time.Time, props ...PropertyParameter) {
	props = append(props, WithValue(string(ValueDataTypeDate)))
	event.SetProperty(ComponentPropertyDtEnd, t.Format(icalDateFormatLocal), props...)
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

func (event *VEvent) getTimeProp(componentProperty ComponentProperty, expectAllDay bool) (time.Time, error) {
	timeProp := event.GetProperty(componentProperty)
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

func (event *VEvent) GetStartAt() (time.Time, error) {
	return event.getTimeProp(ComponentPropertyDtStart, false)
}

func (event *VEvent) GetEndAt() (time.Time, error) {
	return event.getTimeProp(ComponentPropertyDtEnd, false)
}

func (event *VEvent) GetAllDayStartAt() (time.Time, error) {
	return event.getTimeProp(ComponentPropertyDtStart, true)
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

func (event *VEvent) SetSummary(s string, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertySummary, ToText(s), props...)
}

func (event *VEvent) SetStatus(s ObjectStatus, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyStatus, ToText(string(s)), props...)
}

func (event *VEvent) SetDescription(s string, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyDescription, ToText(s), props...)
}

func (event *VEvent) SetLocation(s string, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyLocation, ToText(s), props...)
}

func (event *VEvent) SetGeo(lat interface{}, lng interface{}, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyGeo, fmt.Sprintf("%v;%v", lat, lng), props...)
}

func (event *VEvent) SetURL(s string, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyUrl, s, props...)
}

func (event *VEvent) SetOrganizer(s string, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyOrganizer, s, props...)
}

func (event *VEvent) SetColor(s string, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyColor, s, props...)
}

func (event *VEvent) SetClass(c Classification, props ...PropertyParameter) {
	event.SetProperty(ComponentPropertyClass, string(c), props...)
}

func (event *VEvent) AddAttendee(s string, props ...PropertyParameter) {
	event.AddProperty(ComponentPropertyAttendee, "mailto:"+s, props...)
}

func (event *VEvent) AddExdate(s string, props ...PropertyParameter) {
	event.AddProperty(ComponentPropertyExdate, s, props...)
}

func (event *VEvent) AddExrule(s string, props ...PropertyParameter) {
	event.AddProperty(ComponentPropertyExrule, s, props...)
}

func (event *VEvent) AddRdate(s string, props ...PropertyParameter) {
	event.AddProperty(ComponentPropertyRdate, s, props...)
}

func (event *VEvent) AddRrule(s string, props ...PropertyParameter) {
	event.AddProperty(ComponentPropertyRrule, s, props...)
}

func (event *VEvent) AddAttachment(s string, props ...PropertyParameter) {
	event.AddProperty(ComponentPropertyAttach, s, props...)
}

func (event *VEvent) AddAttachmentURL(uri string, contentType string) {
	event.AddAttachment(uri, WithFmtType(contentType))
}

func (event *VEvent) AddAttachmentBinary(binary []byte, contentType string) {
	event.AddAttachment(base64.StdEncoding.EncodeToString(binary),
		WithFmtType(contentType), WithEncoding("base64"), WithValue("binary"),
	)
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

func (event *VEvent) Attendees() (r []*Attendee) {
	r = []*Attendee{}
	for i := range event.Properties {
		switch event.Properties[i].IANAToken {
		case string(ComponentPropertyAttendee):
			a := &Attendee{
				event.Properties[i],
			}
			r = append(r, a)
		}
	}
	return
}

func (event *VEvent) Id() string {
	p := event.GetProperty(ComponentPropertyUniqueId)
	if p != nil {
		return FromText(p.Value)
	}
	return ""
}

func (event *VEvent) AddAlarm() *VAlarm {
	a := &VAlarm{
		ComponentBase: ComponentBase{},
	}
	event.Components = append(event.Components, a)
	return a
}

func (event *VEvent) Alarms() (r []*VAlarm) {
	r = []*VAlarm{}
	for i := range event.Components {
		switch alarm := event.Components[i].(type) {
		case *VAlarm:
			r = append(r, alarm)
		}
	}
	return
}

type VTodo struct {
	ComponentBase
}

func (c *VTodo) serialize(w io.Writer) {
	c.ComponentBase.serializeThis(w, "VTODO")
}

func (c *VTodo) Serialize() string {
	b := &bytes.Buffer{}
	c.ComponentBase.serializeThis(b, "VTODO")
	return b.String()
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
