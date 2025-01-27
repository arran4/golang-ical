package ics

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
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

type ComponentProperty Property

const (
	ComponentPropertyUniqueId        = ComponentProperty(PropertyUid) // TEXT
	ComponentPropertyDtstamp         = ComponentProperty(PropertyDtstamp)
	ComponentPropertyOrganizer       = ComponentProperty(PropertyOrganizer)
	ComponentPropertyAttendee        = ComponentProperty(PropertyAttendee)
	ComponentPropertyAttach          = ComponentProperty(PropertyAttach)
	ComponentPropertyDescription     = ComponentProperty(PropertyDescription) // TEXT
	ComponentPropertyCategories      = ComponentProperty(PropertyCategories)  // TEXT
	ComponentPropertyClass           = ComponentProperty(PropertyClass)       // TEXT
	ComponentPropertyColor           = ComponentProperty(PropertyColor)       // TEXT
	ComponentPropertyCreated         = ComponentProperty(PropertyCreated)
	ComponentPropertySummary         = ComponentProperty(PropertySummary) // TEXT
	ComponentPropertyDtStart         = ComponentProperty(PropertyDtstart)
	ComponentPropertyDtEnd           = ComponentProperty(PropertyDtend)
	ComponentPropertyLocation        = ComponentProperty(PropertyLocation) // TEXT
	ComponentPropertyStatus          = ComponentProperty(PropertyStatus)   // TEXT
	ComponentPropertyFreebusy        = ComponentProperty(PropertyFreebusy)
	ComponentPropertyLastModified    = ComponentProperty(PropertyLastModified)
	ComponentPropertyUrl             = ComponentProperty(PropertyUrl)
	ComponentPropertyGeo             = ComponentProperty(PropertyGeo)
	ComponentPropertyTransp          = ComponentProperty(PropertyTransp)
	ComponentPropertySequence        = ComponentProperty(PropertySequence)
	ComponentPropertyExdate          = ComponentProperty(PropertyExdate)
	ComponentPropertyExrule          = ComponentProperty(PropertyExrule)
	ComponentPropertyRdate           = ComponentProperty(PropertyRdate)
	ComponentPropertyRrule           = ComponentProperty(PropertyRrule)
	ComponentPropertyAction          = ComponentProperty(PropertyAction)
	ComponentPropertyTrigger         = ComponentProperty(PropertyTrigger)
	ComponentPropertyPriority        = ComponentProperty(PropertyPriority)
	ComponentPropertyResources       = ComponentProperty(PropertyResources)
	ComponentPropertyCompleted       = ComponentProperty(PropertyCompleted)
	ComponentPropertyDue             = ComponentProperty(PropertyDue)
	ComponentPropertyPercentComplete = ComponentProperty(PropertyPercentComplete)
	ComponentPropertyTzid            = ComponentProperty(PropertyTzid)
	ComponentPropertyComment         = ComponentProperty(PropertyComment)
	ComponentPropertyRelatedTo       = ComponentProperty(PropertyRelatedTo)
	ComponentPropertyMethod          = ComponentProperty(PropertyMethod)
	ComponentPropertyRecurrenceId    = ComponentProperty(PropertyRecurrenceId)
	ComponentPropertyDuration        = ComponentProperty(PropertyDuration)
	ComponentPropertyContact         = ComponentProperty(PropertyContact)
	ComponentPropertyRequestStatus   = ComponentProperty(PropertyRequestStatus)
	ComponentPropertyRDate           = ComponentProperty(PropertyRdate)
)

// Required returns the rules from the RFC as to if they are required or not for any particular component type
// If unspecified or incomplete, it returns false. -- This list is incomplete verify source. Happy to take PRs with reference
// iana-prop and x-props are not covered as it would always be true and require an exhaustive list.
func (cp ComponentProperty) Required(c Component) bool {
	// https://www.rfc-editor.org/rfc/rfc5545#section-3.6.1
	switch cp {
	case ComponentPropertyDtstamp, ComponentPropertyUniqueId:
		switch c.(type) {
		case *VEvent:
			return true
		}
	case ComponentPropertyDtStart:
		switch c := c.(type) {
		case *VEvent:
			return !c.HasProperty(ComponentPropertyMethod)
		}
	}
	return false
}

// Exclusive returns the ComponentProperty's using the rules from the RFC as to if one or more existing properties are prohibiting this one
// If unspecified or incomplete, it returns false. -- This list is incomplete verify source. Happy to take PRs with reference
// iana-prop and x-props are not covered as it would always be true and require an exhaustive list.
func (cp ComponentProperty) Exclusive(c Component) []ComponentProperty {
	// https://www.rfc-editor.org/rfc/rfc5545#section-3.6.1
	switch cp {
	case ComponentPropertyDtEnd:
		switch c := c.(type) {
		case *VEvent:
			if c.HasProperty(ComponentPropertyDuration) {
				return []ComponentProperty{ComponentPropertyDuration}
			}
		}
	case ComponentPropertyDuration:
		switch c := c.(type) {
		case *VEvent:
			if c.HasProperty(ComponentPropertyDtEnd) {
				return []ComponentProperty{ComponentPropertyDtEnd}
			}
		}
	}
	return nil
}

// Singular returns the rules from the RFC as to if the spec states that if "Must not occur more than once"
// iana-prop and x-props are not covered as it would always be true and require an exhaustive list.
func (cp ComponentProperty) Singular(c Component) bool {
	// https://www.rfc-editor.org/rfc/rfc5545#section-3.6.1
	switch cp {
	case ComponentPropertyClass, ComponentPropertyCreated, ComponentPropertyDescription, ComponentPropertyGeo,
		ComponentPropertyLastModified, ComponentPropertyLocation, ComponentPropertyOrganizer, ComponentPropertyPriority,
		ComponentPropertySequence, ComponentPropertyStatus, ComponentPropertySummary, ComponentPropertyTransp,
		ComponentPropertyUrl, ComponentPropertyRecurrenceId:
		switch c.(type) {
		case *VEvent:
			return true
		}
	}
	return false
}

// Optional returns the rules from the RFC as to if the spec states that if these are optional
// iana-prop and x-props are not covered as it would always be true and require an exhaustive list.
func (cp ComponentProperty) Optional(c Component) bool {
	// https://www.rfc-editor.org/rfc/rfc5545#section-3.6.1
	switch cp {
	case ComponentPropertyClass, ComponentPropertyCreated, ComponentPropertyDescription, ComponentPropertyGeo,
		ComponentPropertyLastModified, ComponentPropertyLocation, ComponentPropertyOrganizer, ComponentPropertyPriority,
		ComponentPropertySequence, ComponentPropertyStatus, ComponentPropertySummary, ComponentPropertyTransp,
		ComponentPropertyUrl, ComponentPropertyRecurrenceId, ComponentPropertyRrule, ComponentPropertyAttach,
		ComponentPropertyAttendee, ComponentPropertyCategories, ComponentPropertyComment,
		ComponentPropertyContact, ComponentPropertyExdate, ComponentPropertyRequestStatus, ComponentPropertyRelatedTo,
		ComponentPropertyResources, ComponentPropertyRDate:
		switch c.(type) {
		case *VEvent:
			return true
		}
	}
	return false
}

// Multiple returns the rules from the RFC as to if the spec states explicitly if multiple are allowed
// iana-prop and x-props are not covered as it would always be true and require an exhaustive list.
func (cp ComponentProperty) Multiple(c Component) bool {
	// https://www.rfc-editor.org/rfc/rfc5545#section-3.6.1
	switch cp {
	case ComponentPropertyAttach, ComponentPropertyAttendee, ComponentPropertyCategories, ComponentPropertyComment,
		ComponentPropertyContact, ComponentPropertyExdate, ComponentPropertyRequestStatus, ComponentPropertyRelatedTo,
		ComponentPropertyResources, ComponentPropertyRDate:
		switch c.(type) {
		case *VEvent:
			return true
		}
	}
	return false
}

func ComponentPropertyExtended(s string) ComponentProperty {
	return ComponentProperty("X-" + strings.TrimPrefix("X-", s))
}

type Property string

const (
	PropertyCalscale        Property = "CALSCALE" // TEXT
	PropertyMethod          Property = "METHOD"   // TEXT
	PropertyProductId       Property = "PRODID"   // TEXT
	PropertyVersion         Property = "VERSION"  // TEXT
	PropertyXPublishedTTL   Property = "X-PUBLISHED-TTL"
	PropertyRefreshInterval Property = "REFRESH-INTERVAL;VALUE=DURATION"
	PropertyAttach          Property = "ATTACH"
	PropertyCategories      Property = "CATEGORIES"  // TEXT
	PropertyClass           Property = "CLASS"       // TEXT
	PropertyColor           Property = "COLOR"       // TEXT
	PropertyComment         Property = "COMMENT"     // TEXT
	PropertyDescription     Property = "DESCRIPTION" // TEXT
	PropertyXWRCalDesc      Property = "X-WR-CALDESC"
	PropertyGeo             Property = "GEO"
	PropertyLocation        Property = "LOCATION" // TEXT
	PropertyPercentComplete Property = "PERCENT-COMPLETE"
	PropertyPriority        Property = "PRIORITY"
	PropertyResources       Property = "RESOURCES" // TEXT
	PropertyStatus          Property = "STATUS"    // TEXT
	PropertySummary         Property = "SUMMARY"   // TEXT
	PropertyCompleted       Property = "COMPLETED"
	PropertyDtend           Property = "DTEND"
	PropertyDue             Property = "DUE"
	PropertyDtstart         Property = "DTSTART"
	PropertyDuration        Property = "DURATION"
	PropertyFreebusy        Property = "FREEBUSY"
	PropertyTransp          Property = "TRANSP" // TEXT
	PropertyTzid            Property = "TZID"   // TEXT
	PropertyTzname          Property = "TZNAME" // TEXT
	PropertyTzoffsetfrom    Property = "TZOFFSETFROM"
	PropertyTzoffsetto      Property = "TZOFFSETTO"
	PropertyTzurl           Property = "TZURL"
	PropertyAttendee        Property = "ATTENDEE"
	PropertyContact         Property = "CONTACT" // TEXT
	PropertyOrganizer       Property = "ORGANIZER"
	PropertyRecurrenceId    Property = "RECURRENCE-ID"
	PropertyRelatedTo       Property = "RELATED-TO" // TEXT
	PropertyUrl             Property = "URL"
	PropertyUid             Property = "UID" // TEXT
	PropertyExdate          Property = "EXDATE"
	PropertyExrule          Property = "EXRULE"
	PropertyRdate           Property = "RDATE"
	PropertyRrule           Property = "RRULE"
	PropertyAction          Property = "ACTION" // TEXT
	PropertyRepeat          Property = "REPEAT"
	PropertyTrigger         Property = "TRIGGER"
	PropertyCreated         Property = "CREATED"
	PropertyDtstamp         Property = "DTSTAMP"
	PropertyLastModified    Property = "LAST-MODIFIED"
	PropertyRequestStatus   Property = "REQUEST-STATUS" // TEXT
	PropertyName            Property = "NAME"
	PropertyXWRCalName      Property = "X-WR-CALNAME"
	PropertyXWRTimezone     Property = "X-WR-TIMEZONE"
	PropertySequence        Property = "SEQUENCE"
	PropertyXWRCalID        Property = "X-WR-RELCALID"
	PropertyTimezoneId      Property = "TIMEZONE-ID"
)

type Parameter string

func (p Parameter) IsQuoted() bool {
	switch p {
	case ParameterAltrep:
		return true
	}
	return false
}

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

func (cut CalendarUserType) KeyValue(_ ...interface{}) (string, []string) {
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

func (ps ParticipationStatus) KeyValue(_ ...interface{}) (string, []string) {
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

func (ps ObjectStatus) KeyValue(_ ...interface{}) (string, []string) {
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

func (pr ParticipationRole) KeyValue(_ ...interface{}) (string, []string) {
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
	return NewCalendarFor("arran4")
}

func NewCalendarFor(service string) *Calendar {
	c := &Calendar{
		Components:         []Component{},
		CalendarProperties: []CalendarProperty{},
	}
	c.SetVersion("2.0")
	c.SetProductId("-//" + service + "//Golang ICS Library")
	return c
}

func (cal *Calendar) Serialize(ops ...any) string {
	b := &strings.Builder{}
	// We are intentionally ignoring the return value. _ used to communicate this to lint.
	_ = cal.SerializeTo(b, ops...)
	return b.String()
}

type WithLineLength int
type WithNewLine string

func (cal *Calendar) SerializeTo(w io.Writer, ops ...any) error {
	serializeConfig, err := parseSerializeOps(ops)
	if err != nil {
		return err
	}
	_, _ = io.WriteString(w, "BEGIN:VCALENDAR"+serializeConfig.NewLine)
	for _, p := range cal.CalendarProperties {
		err := p.serialize(w, serializeConfig)
		if err != nil {
			return err
		}
	}
	for _, c := range cal.Components {
		err := c.SerializeTo(w, serializeConfig)
		if err != nil {
			return err
		}
	}
	_, _ = io.WriteString(w, "END:VCALENDAR"+serializeConfig.NewLine)
	return nil
}

type SerializationConfiguration struct {
	MaxLength         int
	NewLine           string
	PropertyMaxLength int
}

func parseSerializeOps(ops []any) (*SerializationConfiguration, error) {
	serializeConfig := defaultSerializationOptions()
	for opi, op := range ops {
		switch op := op.(type) {
		case WithLineLength:
			serializeConfig.MaxLength = int(op)
		case WithNewLine:
			serializeConfig.NewLine = string(op)
		case *SerializationConfiguration:
			return op, nil
		case error:
			return nil, op
		default:
			return nil, fmt.Errorf("unknown op %d of type %s", opi, reflect.TypeOf(op))
		}
	}
	return serializeConfig, nil
}

func defaultSerializationOptions() *SerializationConfiguration {
	serializeConfig := &SerializationConfiguration{
		MaxLength:         75,
		PropertyMaxLength: 75,
		NewLine:           string(NewLine),
	}
	return serializeConfig
}

func (cal *Calendar) SetMethod(method Method, params ...PropertyParameter) {
	cal.setProperty(PropertyMethod, string(method), params...)
}

func (cal *Calendar) SetXPublishedTTL(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyXPublishedTTL, s, params...)
}

func (cal *Calendar) SetVersion(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyVersion, s, params...)
}

func (cal *Calendar) SetProductId(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyProductId, s, params...)
}

func (cal *Calendar) SetName(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyName, s, params...)
	cal.setProperty(PropertyXWRCalName, s, params...)
}

func (cal *Calendar) SetColor(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyColor, s, params...)
}

func (cal *Calendar) SetXWRCalName(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyXWRCalName, s, params...)
}

func (cal *Calendar) SetXWRCalDesc(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyXWRCalDesc, s, params...)
}

func (cal *Calendar) SetXWRTimezone(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyXWRTimezone, s, params...)
}

func (cal *Calendar) SetXWRCalID(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyXWRCalID, s, params...)
}

func (cal *Calendar) SetDescription(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyDescription, s, params...)
}

func (cal *Calendar) SetLastModified(t time.Time, params ...PropertyParameter) {
	cal.setProperty(PropertyLastModified, t.UTC().Format(icalTimestampFormatUtc), params...)
}

func (cal *Calendar) SetRefreshInterval(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyRefreshInterval, s, params...)
}

func (cal *Calendar) SetCalscale(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyCalscale, s, params...)
}

func (cal *Calendar) SetUrl(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyUrl, s, params...)
}

func (cal *Calendar) SetTzid(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyTzid, s, params...)
}

func (cal *Calendar) SetTimezoneId(s string, params ...PropertyParameter) {
	cal.setProperty(PropertyTimezoneId, s, params...)
}

func (cal *Calendar) setProperty(property Property, value string, params ...PropertyParameter) {
	for i := range cal.CalendarProperties {
		if cal.CalendarProperties[i].IANAToken == string(property) {
			cal.CalendarProperties[i].Value = value
			cal.CalendarProperties[i].ICalParameters = map[string][]string{}
			for _, p := range params {
				k, v := p.KeyValue()
				cal.CalendarProperties[i].ICalParameters[k] = v
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
	for _, p := range params {
		k, v := p.KeyValue()
		r.ICalParameters[k] = v
	}
	cal.CalendarProperties = append(cal.CalendarProperties, r)
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

func (calendar *Calendar) RemoveEvent(id string) {
	for i := range calendar.Components {
		switch event := calendar.Components[i].(type) {
		case *VEvent:
			if event.Id() == id {
				if len(calendar.Components) > i+1 {
					calendar.Components = append(calendar.Components[:i], calendar.Components[i+1:]...)
				} else {
					calendar.Components = calendar.Components[:i]
				}
				return
			}
		}
	}
}

func WithCustomClient(client *http.Client) *http.Client {
	return client
}

func WithCustomRequest(request *http.Request) *http.Request {
	return request
}

func ParseCalendarFromUrl(url string, opts ...any) (*Calendar, error) {
	var ctx context.Context
	var req *http.Request
	var client HttpClientLike = http.DefaultClient
	for opti, opt := range opts {
		switch opt := opt.(type) {
		case *http.Client:
			client = opt
		case HttpClientLike:
			client = opt
		case func() *http.Client:
			client = opt()
		case *http.Request:
			req = opt
		case func() *http.Request:
			req = opt()
		case context.Context:
			ctx = opt
		case func() context.Context:
			ctx = opt()
		default:
			return nil, fmt.Errorf("unknown optional argument %d on ParseCalendarFromUrl: %s", opti, reflect.TypeOf(opt))
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if req == nil {
		var err error
		req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating http request: %w", err)
		}
	}
	return parseCalendarFromHttpRequest(client, req)
}

type HttpClientLike interface {
	Do(req *http.Request) (*http.Response, error)
}

func parseCalendarFromHttpRequest(client HttpClientLike, request *http.Request) (*Calendar, error) {
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer func(closer io.ReadCloser) {
		if derr := closer.Close(); derr != nil && err == nil {
			err = fmt.Errorf("http request close: %w", derr)
		}
	}(resp.Body)
	var cal *Calendar
	cal, err = ParseCalendar(resp.Body)
	// This allows the defer func to change the error
	return cal, err
}

func ParseCalendar(r io.Reader) (*Calendar, error) {
	state := "begin"
	c := &Calendar{}
	cs := NewCalendarStream(r)
	cont := true
	for ln := 0; cont; ln++ {
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
		line, err := ParseProperty(*l)
		if err != nil {
			return nil, fmt.Errorf("parsing line %d: %w", ln, err)
		}
		if line == nil {
			return nil, fmt.Errorf("parsing calendar line %d", ln)
		}
		switch state {
		case "begin":
			switch line.IANAToken {
			case "BEGIN":
				switch line.Value {
				case "VCALENDAR":
					state = "properties"
				default:
					return nil, errors.New("malformed calendar; expected a vcalendar")
				}
			default:
				return nil, errors.New("malformed calendar; expected begin")
			}
		case "properties":
			switch line.IANAToken {
			case "END":
				switch line.Value {
				case "VCALENDAR":
					state = "end"
				default:
					return nil, errors.New("malformed calendar; expected end")
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
					return nil, errors.New("malformed calendar; expected end")
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
				return nil, errors.New("malformed calendar; expected begin or end")
			}
		case "end":
			return nil, errors.New("malformed calendar; unexpected end")
		default:
			return nil, errors.New("malformed calendar; bad state")
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
		switch {
		case len(b) == 0:
			if err == nil {
				continue
			} else {
				c = false
			}
		case b[len(b)-1] == '\n':
			o := 1
			if len(b) > 1 && b[len(b)-2] == '\r' {
				o = 2
			}
			p, err := cs.b.Peek(1)
			r = append(r, b[:len(b)-o]...)
			if err == io.EOF {
				c = false
			}
			switch {
			case len(p) == 0:
				c = false
			case p[0] == ' ' || p[0] == '\t':
				_, _ = cs.b.Discard(1) // nolint:errcheck
			default:
				c = false
			}
		default:
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
