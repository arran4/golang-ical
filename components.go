package ics

import (
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

// ComponentBase provides common storage for properties and subcomponents.
// It is embedded in the concrete component structs such as VEvent and VTodo.
// Each component begins with a "BEGIN" content line and ends with "END" as
// defined in RFC 5545 section 3.4.
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
	_, _ = io.WriteString(writer, "BEGIN:"+string(componentType)+serialConfig.NewLine)
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
	_, err := io.WriteString(writer, "END:"+string(componentType)+serialConfig.NewLine)
	return err
}

// NewComponent creates a minimal component with the required UID property.
// RFC 5545 section 3.8.4.7 states that the UID uniquely identifies the
// component.  Callers typically embed the returned ComponentBase into a
// specific component struct.
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
	for i := range cb.Properties {
		if cb.Properties[i].IANAToken == string(componentProperty) {
			return &cb.Properties[i]
		}
	}
	return nil
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

// The following layout constants format DATE and DATE-TIME values according to
// RFC 5545 section 3.3.
const (
	// icalTimestampFormatUtc is the layout for DATE-TIME in UTC.
	icalTimestampFormatUtc = "20060102T150405Z"
	// icalTimestampFormatLocal is the layout for DATE-TIME without timezone.
	icalTimestampFormatLocal = "20060102T150405"
	// icalDateFormatUtc is the layout for DATE in UTC.
	icalDateFormatUtc = "20060102Z"
	// icalDateFormatLocal is the layout for DATE in local time.
	icalDateFormatLocal = "20060102"
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

// SetDuration updates the duration of an event without explicitly setting the
// DURATION property.  RFC 5545 section 3.3.6 states "The DURATION value type
// expresses a period of time."  This helper adjusts DTSTART or DTEND so the
// resulting period equals d.  If both a start and end are present, one is
// modified to satisfy the requested duration.
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
		return time.Time{}, fmt.Errorf("%w: %s", ErrorPropertyNotFound, componentProperty)
	}

	timeVal := timeProp.Value
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

func (cb *ComponentBase) SetSummary(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertySummary, s, params...)
}

func (cb *ComponentBase) SetStatus(s ObjectStatus, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyStatus, string(s), params...)
}

// SetDescription updates the DESCRIPTION property on the component.
// RFC 5545 section 3.8.1.5 explains this text supplies descriptive
// information that calendar clients typically display to the user.
func (cb *ComponentBase) SetDescription(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyDescription, s, params...)
}

// SetLocation stores the LOCATION property (RFC 5545 section 3.8.1.7)
// describing where the event will take place.
func (cb *ComponentBase) SetLocation(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyLocation, s, params...)
}

func (cb *ComponentBase) setGeo(lat interface{}, lng interface{}, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyGeo, fmt.Sprintf("%v;%v", lat, lng), params...)
}

// SetURL sets the URL property as defined in RFC 5545 section 3.8.4.6.
// The URL points to additional information about the component.
func (cb *ComponentBase) SetURL(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyUrl, s, params...)
}

// SetOrganizer sets the ORGANIZER property (RFC 5545 section 3.8.4.3)
// to the specified calendar address. The address is prefixed with
// "mailto:" if not already present.
func (cb *ComponentBase) SetOrganizer(s string, params ...PropertyParameter) {
	if !strings.HasPrefix(s, "mailto:") {
		s = "mailto:" + s
	}
	cb.SetProperty(ComponentPropertyOrganizer, s, params...)
}

// SetColor stores the non-standard COLOR property which many clients
// use to style calendar entries.
func (cb *ComponentBase) SetColor(s string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyColor, s, params...)
}

// SetClass sets the CLASS property controlling access to the
// component as described in RFC 5545 section 3.8.1.3.
func (cb *ComponentBase) SetClass(c Classification, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyClass, string(c), params...)
}

func (cb *ComponentBase) setPriority(p int, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyPriority, strconv.Itoa(p), params...)
}

func (cb *ComponentBase) setResources(r string, params ...PropertyParameter) {
	cb.SetProperty(ComponentPropertyResources, r, params...)
}

// AddAttendee appends an ATTENDEE property (RFC 5545 section 3.8.4.1)
// identifying a participant of the event.
func (cb *ComponentBase) AddAttendee(s string, params ...PropertyParameter) {
	if !strings.HasPrefix(s, "mailto:") {
		s = "mailto:" + s
	}
	cb.AddProperty(ComponentPropertyAttendee, s, params...)
}

// AddExdate adds an EXDATE property (RFC 5545 section 3.8.5.1)
// excluding the specified recurrence date.
func (cb *ComponentBase) AddExdate(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyExdate, s, params...)
}

// AddExrule appends the deprecated EXRULE property. It remains for
// backward compatibility with RFC 2445 calendars.
func (cb *ComponentBase) AddExrule(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyExrule, s, params...)
}

// AddRdate appends an RDATE property containing an additional
// recurrence date (RFC 5545 section 3.8.5.2).
func (cb *ComponentBase) AddRdate(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyRdate, s, params...)
}

// AddRrule appends an RRULE property (RFC 5545 section 3.8.5.3)
// defining the recurrence rule for the component.
func (cb *ComponentBase) AddRrule(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyRrule, s, params...)
}

// AddAttachment appends an ATTACH property (RFC 5545 section 3.8.1.1).
// Use AddAttachmentURL or AddAttachmentBinary for convenience.
func (cb *ComponentBase) AddAttachment(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyAttach, s, params...)
}

// AddAttachmentURL is a helper that records an attachment by URL with
// its associated MIME type.
func (cb *ComponentBase) AddAttachmentURL(uri string, contentType string) {
	cb.AddAttachment(uri, WithFmtType(contentType))
}

// AddAttachmentBinary attaches raw binary data to the component. The
// data is base64 encoded and stored in an ATTACH property.
func (cb *ComponentBase) AddAttachmentBinary(binary []byte, contentType string) {
	cb.AddAttachment(base64.StdEncoding.EncodeToString(binary),
		WithFmtType(contentType), WithEncoding("base64"), WithValue("binary"),
	)
}

// AddComment appends a COMMENT property (RFC 5545 section 3.8.1.4).
func (cb *ComponentBase) AddComment(s string, params ...PropertyParameter) {
	cb.AddProperty(ComponentPropertyComment, s, params...)
}

// AddCategory adds a CATEGORIES property which can be repeated to
// classify the component (RFC 5545 section 3.8.1.2).
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

func (event *VEvent) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return event.serializeThis(w, ComponentVEvent, serialConfig)
}

func (event *VEvent) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := event.serialize(serialConfig)
	return s
}

func (event *VEvent) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &strings.Builder{}
	err := event.serializeThis(b, ComponentVEvent, serialConfig)
	return b.String(), err
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

// SetGeo records the geographic position of the event.  Latitude and longitude
// may be provided as any numeric type.  Values are formatted using
// fmt.Sprintf and stored in the GEO property (RFC 5545 section 3.8.1.6).
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

// TimeTransparency enumerates values for the TRANSP property (RFC 5545 section 3.8.2.7).
const (
	// TransparencyOpaque marks busy time that blocks availability.
	TransparencyOpaque TimeTransparency = "OPAQUE" // default
	// TransparencyTransparent marks busy time that does not block availability.
	TransparencyTransparent TimeTransparency = "TRANSPARENT"
)

func (event *VEvent) SetTimeTransparency(v TimeTransparency, params ...PropertyParameter) {
	event.SetProperty(ComponentPropertyTransp, string(v), params...)
}

type VTodo struct {
	ComponentBase
}

func (todo *VTodo) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return todo.serializeThis(w, ComponentVTodo, serialConfig)
}

func (todo *VTodo) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := todo.serialize(serialConfig)
	return s
}

func (todo *VTodo) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &strings.Builder{}
	err := todo.serializeThis(b, ComponentVTodo, serialConfig)
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
	todo.SetProperty(ComponentPropertyCompleted, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (todo *VTodo) SetAllDayCompletedAt(t time.Time, params ...PropertyParameter) {
	params = append(params, WithValue(string(ValueDataTypeDate)))
	todo.SetProperty(ComponentPropertyCompleted, t.Format(icalDateFormatLocal), params...)
}

func (todo *VTodo) SetDueAt(t time.Time, params ...PropertyParameter) {
	todo.SetProperty(ComponentPropertyDue, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (todo *VTodo) SetAllDayDueAt(t time.Time, params ...PropertyParameter) {
	params = append(params, WithValue(string(ValueDataTypeDate)))
	todo.SetProperty(ComponentPropertyDue, t.Format(icalDateFormatLocal), params...)
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
// This function mirrors SetDuration on ComponentBase.  It adjusts DTSTART or
// DUE so that the task spans the requested duration as defined by RFC 5545
// section 3.3.6 without adding a DURATION property.
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
	return journal.serializeThis(w, ComponentVJournal, serialConfig)
}

func (journal *VJournal) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := journal.serialize(serialConfig)
	return s
}

func (journal *VJournal) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &strings.Builder{}
	err := journal.serializeThis(b, ComponentVJournal, serialConfig)
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
	b := &strings.Builder{}
	err := busy.serializeThis(b, ComponentVFreeBusy, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (busy *VBusy) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return busy.serializeThis(w, ComponentVFreeBusy, serialConfig)
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
	b := &strings.Builder{}
	err := timezone.serializeThis(b, ComponentVTimezone, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (timezone *VTimezone) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return timezone.serializeThis(w, ComponentVTimezone, serialConfig)
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
	b := &strings.Builder{}
	err := c.serializeThis(b, ComponentVAlarm, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (c *VAlarm) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return c.serializeThis(w, ComponentVAlarm, serialConfig)
}

// NewAlarm creates a bare VALARM component.  The tzId argument is kept for
// backward compatibility but is not currently used when constructing the
// component.
func NewAlarm(tzId string) *VAlarm {
	_ = tzId // parameter reserved for future use
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
	b := &strings.Builder{}
	err := standard.serializeThis(b, ComponentStandard, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (standard *Standard) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return standard.serializeThis(w, ComponentStandard, serialConfig)
}

type Daylight struct {
	ComponentBase
}

func (daylight *Daylight) Serialize(serialConfig *SerializationConfiguration) string {
	s, _ := daylight.serialize(serialConfig)
	return s
}

func (daylight *Daylight) serialize(serialConfig *SerializationConfiguration) (string, error) {
	b := &strings.Builder{}
	err := daylight.serializeThis(b, ComponentDaylight, serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (daylight *Daylight) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return daylight.serializeThis(w, ComponentDaylight, serialConfig)
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
	b := &strings.Builder{}
	err := general.serializeThis(b, ComponentType(general.Token), serialConfig)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (general *GeneralComponent) SerializeTo(w io.Writer, serialConfig *SerializationConfiguration) error {
	return general.serializeThis(w, ComponentType(general.Token), serialConfig)
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
		default:
			// Unrecognised property names are retained verbatim as
			// IANA properties so that non-standard extensions are
			// preserved during round trip parsing and serialization.
			cb.Properties = append(cb.Properties, IANAProperty{*line})
		}
	}
	return cb, errors.New("ran out of lines")
}
