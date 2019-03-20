package ics

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

type ComponentType string

const (
	ComponentVCalendar ComponentType = "VCALENDAR"
	ComponentVEvent    ComponentType = "VEVENT"
	ComponentVTodo     ComponentType = "VTODO"
	ComponentVJournal  ComponentType = "VJOURNAL"
	ComponentVFreeBusy ComponentType = "VFREEBUSY"
	ComponentVTimezone ComponentType = "VTIMEZONE"
	ComponentVAlarm    ComponentType = "VALARM"
	ComponentStandard  ComponentType = "STANDARD"
	ComponentDaylight  ComponentType = "DAYLIGHT"
)

type ComponentProperty string

const (
	ComponentPropertyUniqueId     ComponentProperty = "UID"
	ComponentPropertyDtstamp      ComponentProperty = "DTSTAMP"
	ComponentPropertyOrganizer    ComponentProperty = "ORGANIZER"
	ComponentPropertyAttendee     ComponentProperty = "ATTENDEE"
	ComponentPropertyDescription  ComponentProperty = "DESCRIPTION"
	ComponentPropertyCategories   ComponentProperty = "CATEGORIES"
	ComponentPropertyClass        ComponentProperty = "CLASS"
	ComponentPropertyCreated      ComponentProperty = "CREATED"
	ComponentPropertySummary      ComponentProperty = "SUMMARY"
	ComponentPropertyDtStart      ComponentProperty = "DTSTART"
	ComponentPropertyDtEnd        ComponentProperty = "DTEND"
	ComponentPropertyLocation     ComponentProperty = "LOCATION"
	ComponentPropertyStatus       ComponentProperty = "STATUS"
	ComponentPropertyFreebusy     ComponentProperty = "FREEBUSY"
	ComponentPropertyLastModified ComponentProperty = "LAST-MODIFIED"
	ComponentPropertyUrl          ComponentProperty = "URL"
)

type Property string

const (
	PropertyCalscale        Property = "CALSCALE"
	PropertyMethod          Property = "METHOD"
	PropertyProductId       Property = "PRODID"
	PropertyVersion         Property = "VERSION"
	PropertyAttach          Property = "ATTACH"
	PropertyCategories      Property = "CATEGORIES"
	PropertyClass           Property = "CLASS"
	PropertyComment         Property = "COMMENT"
	PropertyDescription     Property = "DESCRIPTION"
	PropertyGeo             Property = "GEO"
	PropertyLocation        Property = "LOCATION"
	PropertyPercentComplete Property = "PERCENT-COMPLETE"
	PropertyPriority        Property = "PRIORITY"
	PropertyResources       Property = "RESOURCES"
	PropertyStatus          Property = "STATUS"
	PropertySummary         Property = "SUMMARY"
	PropertyCompleted       Property = "COMPLETED"
	PropertyDtend           Property = "DTEND"
	PropertyDue             Property = "DUE"
	PropertyDtstart         Property = "DTSTART"
	PropertyDuration        Property = "DURATION"
	PropertyFreebusy        Property = "FREEBUSY"
	PropertyTransp          Property = "TRANSP"
	PropertyTzid            Property = "TZID"
	PropertyTzname          Property = "TZNAME"
	PropertyTzoffsetfrom    Property = "TZOFFSETFROM"
	PropertyTzoffsetto      Property = "TZOFFSETTO"
	PropertyTzurl           Property = "TZURL"
	PropertyAttendee        Property = "ATTENDEE"
	PropertyContact         Property = "CONTACT"
	PropertyOrganizer       Property = "ORGANIZER"
	PropertyRecurrenceId    Property = "RECURRENCE-ID"
	PropertyRelatedTo       Property = "RELATED-TO"
	PropertyUrl             Property = "URL"
	PropertyUid             Property = "UID"
	PropertyExdate          Property = "EXDATE"
	PropertyExrule          Property = "EXRULE"
	PropertyRdate           Property = "RDATE"
	PropertyRrule           Property = "RRULE"
	PropertyAction          Property = "ACTION"
	PropertyRepeat          Property = "REPEAT"
	PropertyTrigger         Property = "TRIGGER"
	PropertyCreated         Property = "CREATED"
	PropertyDtstamp         Property = "DTSTAMP"
	PropertyLastModified    Property = "LAST-MODIFIED"
	PropertySequence        Property = "SEQUENCE"
	PropertyRequestStatus   Property = "REQUEST-STATUS"
)

type Parameter string

const (
	ParameterAltrep              Parameter = "ALTREP"
	ParameterCn                  Parameter = "CN"
	ParameterCutype              Parameter = "CUTYPE"
	ParameterDelegatedFrom       Parameter = "DELEGATED-FROM"
	ParameterDelegatedTo         Parameter = "DELEGATED-TO"
	ParameterDir                 Parameter = "DIR"
	ParameterEncoding            Parameter = "ENCODING"
	ParameterFmttype             Parameter = "FMTTYPE"
	ParameterFbtype              Parameter = "FBTYPE"
	ParameterLanguage            Parameter = "LANGUAGE"
	ParameterMember              Parameter = "MEMBER"
	ParameterParticipationStatus Parameter = "PARTSTAT"
	ParameterRange               Parameter = "RANGE"
	ParameterRelated             Parameter = "RELATED"
	ParameterReltype             Parameter = "RELTYPE"
	ParameterRole                Parameter = "ROLE"
	ParameterRsvp                Parameter = "RSVP"
	ParameterSentBy              Parameter = "SENT-BY"
	ParameterTzid                Parameter = "TZID"
	ParameterValue               Parameter = "VALUE"
)

type ValueDataType string

const (
	ValueDataTypeBinary     ValueDataType = "BINARY"
	ValueDataTypeBoolean    ValueDataType = "BOOLEAN"
	ValueDataTypeCalAddress ValueDataType = "CAL-ADDRESS"
	ValueDataTypeDate       ValueDataType = "DATE"
	ValueDataTypeDateTime   ValueDataType = "DATE-TIME"
	ValueDataTypeDuration   ValueDataType = "DURATION"
	ValueDataTypeFloat      ValueDataType = "FLOAT"
	ValueDataTypeInteger    ValueDataType = "INTEGER"
	ValueDataTypePeriod     ValueDataType = "PERIOD"
	ValueDataTypeRecur      ValueDataType = "RECUR"
	ValueDataTypeText       ValueDataType = "TEXT"
	ValueDataTypeTime       ValueDataType = "TIME"
	ValueDataTypeUri        ValueDataType = "URI"
	ValueDataTypeUtcOffset  ValueDataType = "UTC-OFFSET"
)

type CalendarUserType string

const (
	CalendarUserTypeIndividual CalendarUserType = "INDIVIDUAL"
	CalendarUserTypeGroup      CalendarUserType = "GROUP"
	CalendarUserTypeResource   CalendarUserType = "RESOURCE"
	CalendarUserTypeRoom       CalendarUserType = "ROOM"
	CalendarUserTypeUnknown    CalendarUserType = "UNKNOWN"
)

func (cut CalendarUserType) KeyValue(s ...interface{}) (string, []string) {
	return string(ParameterCutype), []string{string(cut)}
}

type FreeBusyTimeType string

const (
	FreeBusyTimeTypeFree            FreeBusyTimeType = "FREE"
	FreeBusyTimeTypeBusy            FreeBusyTimeType = "BUSY"
	FreeBusyTimeTypeBusyUnavailable FreeBusyTimeType = "BUSY-UNAVAILABLE"
	FreeBusyTimeTypeBusyTentative   FreeBusyTimeType = "BUSY-TENTATIVE"
)

type ParticipationStatus string

const (
	ParticipationStatusNeedsAction ParticipationStatus = "NEEDS-ACTION"
	ParticipationStatusAccepted    ParticipationStatus = "ACCEPTED"
	ParticipationStatusDeclined    ParticipationStatus = "DECLINED"
	ParticipationStatusTentative   ParticipationStatus = "TENTATIVE"
	ParticipationStatusDelegated   ParticipationStatus = "DELEGATED"
	ParticipationStatusCompleted   ParticipationStatus = "COMPLETED"
	ParticipationStatusInProcess   ParticipationStatus = "IN-PROCESS"
)

func (ps ParticipationStatus) KeyValue(s ...interface{}) (string, []string) {
	return string(ParameterParticipationStatus), []string{string(ps)}
}

type ObjectStatus string

const (
	ObjectStatusTentative   ObjectStatus = "TENTATIVE"
	ObjectStatusConfirmed   ObjectStatus = "CONFIRMED"
	ObjectStatusCancelled   ObjectStatus = "CANCELLED"
	ObjectStatusNeedsAction ObjectStatus = "NEEDS-ACTION"
	ObjectStatusCompleted   ObjectStatus = "COMPLETED"
	ObjectStatusInProcess   ObjectStatus = "IN-PROCESS"
	ObjectStatusDraft       ObjectStatus = "DRAFT"
	ObjectStatusFinal       ObjectStatus = "FINAL"
)

func (ps ObjectStatus) KeyValue(s ...interface{}) (string, []string) {
	return string(PropertyStatus), []string{string(ps)}
}

type RelationshipType string

const (
	RelationshipTypeChild   RelationshipType = "CHILD"
	RelationshipTypeParent  RelationshipType = "PARENT"
	RelationshipTypeSibling RelationshipType = "SIBLING"
)

type ParticipationRole string

const (
	ParticipationRoleChair          ParticipationRole = "CHAIR"
	ParticipationRoleReqParticipant ParticipationRole = "REQ-PARTICIPANT"
	ParticipationRoleOptParticipant ParticipationRole = "OPT-PARTICIPANT"
	ParticipationRoleNonParticipant ParticipationRole = "NON-PARTICIPANT"
)

func (pr ParticipationRole) KeyValue(s ...interface{}) (string, []string) {
	return string(ParameterRole), []string{string(pr)}
}

type Action string

const (
	ActionAudio     Action = "AUDIO"
	ActionDisplay   Action = "DISPLAY"
	ActionEmail     Action = "EMAIL"
	ActionProcedure Action = "PROCEDURE"
)

type Classification string

const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationPrivate      Classification = "PRIVATE"
	ClassificationConfidential Classification = "CONFIDENTIAL"
)

type Method string

const (
	MethodPublish        Method = "PUBLISH"
	MethodRequest        Method = "REQUEST"
	MethodReply          Method = "REPLY"
	MethodAdd            Method = "ADD"
	MethodCancel         Method = "CANCEL"
	MethodRefresh        Method = "REFRESH"
	MethodCounter        Method = "COUNTER"
	MethodDeclinecounter Method = "DECLINECOUNTER"
)

type CalendarProperty struct {
	BaseProperty
}

type Calendar struct {
	Components         []Component
	CalendarProperties []CalendarProperty
}

func NewCalendar() *Calendar {
	c := &Calendar{
		Components:         []Component{},
		CalendarProperties: []CalendarProperty{},
	}
	c.SetVersion("2.0")
	c.SetProductId("-//Arran Ubels//Golang ICS library")
	return c
}

func (calendar *Calendar) Serialize() string {
	b := bytes.NewBufferString("")
	fmt.Fprintln(b, "BEGIN:VCALENDAR")
	for _, p := range calendar.CalendarProperties {
		p.serialize(b)
	}
	for _, c := range calendar.Components {
		c.serialize(b)
	}
	fmt.Fprintln(b, "END:VCALENDAR")
	return b.String()
}

func (calendar *Calendar) SetMethod(method Method, props ...PropertyParameter) {
	calendar.setProperty(PropertyMethod, string(method), props...)
}

func (calendar *Calendar) SetVersion(s string, props ...PropertyParameter) {
	calendar.setProperty(PropertyVersion, string(s), props...)
}

func (calendar *Calendar) SetProductId(s string, props ...PropertyParameter) {
	calendar.setProperty(PropertyProductId, string(s), props...)
}

func (calendar *Calendar) setProperty(property Property, value string, props ...PropertyParameter) {
	for i := range calendar.CalendarProperties {
		if calendar.CalendarProperties[i].IANAToken == string(property) {
			calendar.CalendarProperties[i].Value = value
			calendar.CalendarProperties[i].ICalParameters = map[string][]string{}
			for _, p := range props {
				k, v := p.KeyValue()
				calendar.CalendarProperties[i].ICalParameters[k] = v
			}
			return
		}
	}
	r := CalendarProperty{
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
	calendar.CalendarProperties = append(calendar.CalendarProperties, r)
}
func (calendar *Calendar) AddEvent(id string) *VEvent {
	e := &VEvent{
		ComponentBase{
			Properties: []IANAProperty{
				IANAProperty{BaseProperty{IANAToken: string(ComponentPropertyUniqueId), Value: id}},
			},
		},
	}
	calendar.Components = append(calendar.Components, e)
	return e
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

func ParseCalendar(r io.Reader) (*Calendar, error) {
	state := "begin"
	c := &Calendar{}
	cs := NewCalendarStream(r)
	cont := true
	for i := 0; cont; i++ {
		l, err := cs.ReadLine()
		if err != nil {
			switch err {
			case io.EOF:
				cont = false
			default:
				return c, err
			}
		}
		if l == nil || len(*l) == 0 {
			continue
		}
		line := ParseProperty(*l)
		if line == nil {
			return nil, errors.New("Error parsing line")
		}
		switch state {
		case "begin":
			switch line.IANAToken {
			case "BEGIN":
				switch line.Value {
				case "VCALENDAR":
					state = "properties"
				default:
					return nil, errors.New("Malformed calendar")
				}
			default:
				return nil, errors.New("Malformed calendar")
			}
		case "properties":
			switch line.IANAToken {
			case "END":
				switch line.Value {
				case "VCALENDAR":
					state = "end"
				default:
					return nil, errors.New("Malformed calendar")
				}
			case "BEGIN":
				state = "components"
			default: // TODO put in all the supported types for type switching etc.
				c.CalendarProperties = append(c.CalendarProperties, CalendarProperty{*line})
			}
			if state != "components" {
				break
			}
			fallthrough
		case "components":
			switch line.IANAToken {
			case "END":
				switch line.Value {
				case "VCALENDAR":
					state = "end"
				default:
					return nil, errors.New("Malformed calendar")
				}
			case "BEGIN":
				co, err := GeneralParseComponent(cs, line)
				if err != nil {
					return nil, err
				}
				if co != nil {
					c.Components = append(c.Components, co)
				}
			default:
				return nil, errors.New("Malformed calendar")
			}
		case "end":
			return nil, errors.New("Malformed calendar")
		default:
			return nil, errors.New("Bad state")
		}
	}
	return c, nil
}

type CalendarStream struct {
	r io.Reader
	b *bufio.Reader
}

func NewCalendarStream(r io.Reader) *CalendarStream {
	return &CalendarStream{
		r: r,
		b: bufio.NewReader(r),
	}
}

func (cs *CalendarStream) ReadLine() (*ContentLine, error) {
	r := []byte{}
	c := true
	var err error
	for c {
		var b []byte
		b, err = cs.b.ReadBytes('\n')
		if b == nil || len(b) == 0 {
			if err == nil {
				continue
			} else {
				c = false
			}
		} else if b[len(b)-1] == '\n' {
			o := 1
			if len(b) > 1 && b[len(b)-2] == '\r' {
				o = 2
			}
			p, err := cs.b.Peek(1)
			r = append(r, b[:len(b)-o]...)
			if err == io.EOF {
				c = false
			}
			if len(p) == 0 {
				c = false
			} else if p[0] == ' ' {
				cs.b.Discard(1)
			} else {
				c = false
			}
		} else {
			r = append(r, b...)
		}
		switch err {
		case nil:
			if len(r) == 0 {
				c = true
			}
		case io.EOF:
			c = false
		default:
			return nil, err
		}
	}
	if len(r) == 0 && err != nil {
		return nil, err
	}
	cl := ContentLine(r)
	return &cl, err
}
