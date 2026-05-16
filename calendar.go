package ics

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	PropertySource          Property = "SOURCE"
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

// Method represents the iCalendar METHOD property.
//
// These should be used with caution when creating simple iCal (.ics) files.
// The iCalendar specification is defined in RFC 5545. It refers to RFC 5546,
// which defines the ADD method as allowing the Organizer to add one or more
// new instances to an existing VEVENT, VTODO, or VJOURNAL using a single iTIP message. The UID
// must be that of the existing event/task/journal.
//
// If you include METHOD: ADD in your .ics file (or use SetMethod(MethodAdd)),
// it is required to refer to existing calendar components. If you are simply
// writing .ics files to import into calendaring tools, it is not likely that
// you will want to use this option. Notably, the Apple Calendar program will
// reject events in .ics files that have this set if they do not refer to
// existing calendar events; some other calendars (like Microsoft Outlook)
// have a more permissive import process and will accept them.
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
	Components                     []Component
	CalendarProperties             []CalendarProperty
	unknownCalendarPropertyHandler func(cal *Calendar, state string, cl *BaseProperty) error
	propertyParser                 PropertyParser
	timezoneMapper                 TimezoneMapper
}

func NewCalendar() *Calendar {
	c, _ := NewCalendarWithOptions(
		WithVersion("2.0"),
		WithProductId("-//arran4//Golang ICS Library"),
	)
	return c
}

func NewCalendarFor(service string) *Calendar {
	c, _ := NewCalendarWithOptions(
		WithVersion("2.0"),
		WithProductId("-//"+service+"//Golang ICS Library"),
	)
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
		err := p.SerializeTo(w, serializeConfig)
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
	timezoneMapper    TimezoneSerializationMapper
}

// SerializationOption provides functional options for Serialize and SerializeTo.
type SerializationOption func(*SerializationConfiguration) error

func parseSerializeOps(ops []any) (*SerializationConfiguration, error) {
	serializeConfig := defaultSerializationOptions()
	for opi, op := range ops {
		switch op := op.(type) {
		case WithLineLength:
			serializeConfig.MaxLength = int(op)
		case WithNewLine:
			serializeConfig.NewLine = string(op)
		case SerializationOption:
			if op != nil {
				if err := op(serializeConfig); err != nil {
					return nil, err
				}
			}
		case *SerializationConfiguration:
			return op, nil
		case TimezoneSerializationMapper:
			serializeConfig.timezoneMapper = op
		case func(*time.Location) (string, bool):
			serializeConfig.timezoneMapper = TimezoneSerializationMapper(op)
		case error:
			return nil, op
		default:
			return nil, fmt.Errorf("%w %d: %T", ErrInvalidOpArg, opi, op)
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

// WithTimezoneMapper configures how Windows timezone identifiers are mapped during parsing.
func WithTimezoneMapper(mapper TimezoneMapper) ParseOption {
	return func(c *Calendar) error {
		c.timezoneMapper = mapper
		return nil
	}
}

// WithWindowsTimezoneMapping enables mapping of Windows timezone names to
// IANA equivalents during calendar parsing.
func WithWindowsTimezoneMapping() ParseOption {
	return WithTimezoneMapper(WindowsTimezoneToIANA)
}

// WithSerializationTimezoneMapper configures how timezone identifiers are mapped during serialization.
func WithSerializationTimezoneMapper(mapper TimezoneSerializationMapper) SerializationOption {
	return func(c *SerializationConfiguration) error {
		c.timezoneMapper = mapper
		return nil
	}
}

// WithWindowsTimezoneMappingForSerialization enables mapping of IANA timezone names
// to Windows equivalents during serialization.
func WithWindowsTimezoneMappingForSerialization() SerializationOption {
	return WithSerializationTimezoneMapper(IANAToWindowsTimezone)
}

// SetMethod sets the METHOD property for the calendar.
//
// These should be used with caution when creating simple iCal (.ics) files.
// The iCalendar specification is defined in RFC 5545. It refers to RFC 5546,
// which defines the ADD method as allowing the Organizer to add one or more
// new instances to an existing VEVENT, VTODO, or VJOURNAL using a single iTIP message. The UID
// must be that of the existing event/task/journal.
//
// If you include METHOD: ADD in your .ics file (or use SetMethod(MethodAdd)),
// it is required to refer to existing calendar components. If you are simply
// writing .ics files to import into calendaring tools, it is not likely that
// you will want to use this option. Notably, the Apple Calendar program will
// reject events in .ics files that have this set if they do not refer to
// existing calendar events; some other calendars (like Microsoft Outlook)
// have a more permissive import process and will accept them.
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

func (cal *Calendar) componentParseOptions() []any {
	opts := make([]any, 0, 2)
	if cal.propertyParser != nil {
		opts = append(opts, cal.propertyParser)
	}
	if cal.timezoneMapper != nil {
		opts = append(opts, cal.timezoneMapper)
	}
	return opts
}

func (cal *Calendar) addComponent(c Component) {
	if c == nil {
		return
	}
	if cal.timezoneMapper != nil {
		if setter, ok := c.(timezoneMapperSetter); ok {
			if getter, ok := c.(timezoneMapperGetter); !ok || getter.getTimezoneMapper() == nil {
				setter.setTimezoneMapper(cal.timezoneMapper)
			}
		}
	}
	cal.Components = append(cal.Components, c)
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
	calendar.addComponent(e)
	return e
}

func (calendar *Calendar) AddVEvent(e *VEvent) {
	calendar.addComponent(e)
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
	parseOpts := make([]any, 0, len(opts))
	for i, opt := range opts {
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
		case ParseOption:
			parseOpts = append(parseOpts, opt)
		case CalendarOption:
			parseOpts = append(parseOpts, opt)
		case PropertyParser:
			parseOpts = append(parseOpts, opt)
		case func(ContentLine) (*BaseProperty, error):
			parseOpts = append(parseOpts, opt)
		case TimezoneMapper:
			parseOpts = append(parseOpts, opt)
		case func(string) *time.Location:
			parseOpts = append(parseOpts, opt)
		default:
			return nil, fmt.Errorf("%w %d: %T", ErrInvalidOpArg, i, opt)
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
	return parseCalendarFromHttpRequest(client, req, parseOpts...)
}

type HttpClientLike interface {
	Do(req *http.Request) (*http.Response, error)
}

func parseCalendarFromHttpRequest(client HttpClientLike, request *http.Request, opts ...any) (*Calendar, error) {
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
	cal, err = ParseCalendarWithOptions(resp.Body, opts...)
	// This allows the defer func to change the error
	return cal, err
}

// ParseOption provides functional options for ParseCalendar
type ParseOption func(*Calendar) error

// CalendarOption provides functional options for Calendar construction.
type CalendarOption func(*Calendar) error

// WithVersion sets the calendar version.
func WithVersion(version string, params ...PropertyParameter) CalendarOption {
	return func(c *Calendar) error {
		c.SetVersion(version, params...)
		return nil
	}
}

// WithProductId sets the calendar product identifier.
func WithProductId(productID string, params ...PropertyParameter) CalendarOption {
	return func(c *Calendar) error {
		c.SetProductId(productID, params...)
		return nil
	}
}

// WithUnknownPropertyHandler allows custom handling of unknown properties
func WithUnknownPropertyHandler(f func(*Calendar, string, *BaseProperty) error) ParseOption {
	return func(c *Calendar) error {
		c.unknownCalendarPropertyHandler = f
		return nil
	}
}

// WithPropertyParser allows custom handling of property parse errors.
// It is a convenience wrapper; ParseCalendarWithOptions also accepts PropertyParser directly.
// When a content line fails to parse (e.g. due to malformed parameter names),
// the parser is called with the raw content line.
//
// The parser can:
//   - Return (*BaseProperty, nil) to use a recovered/replacement property
//   - Return (nil, nil) to skip the property silently
//   - Return (nil, err) to abort parsing with the error
//
// Without this option, any property parse error aborts the entire calendar parse.
// This is useful for real-world ICS feeds that contain non-RFC-compliant properties
// (e.g. parameter names with underscores).
func WithPropertyParser(f PropertyParser) ParseOption {
	return func(c *Calendar) error {
		if f != nil {
			c.propertyParser = f
		}
		return nil
	}
}

// NewCalendarWithOptions constructs a calendar with sane parser defaults and optional overrides.
func NewCalendarWithOptions(options ...any) (*Calendar, error) {
	c := &Calendar{
		Components:                     []Component{},
		CalendarProperties:             []CalendarProperty{},
		unknownCalendarPropertyHandler: DefaultUnknownCalendarPropertyHandler,
		propertyParser:                 parseProperty,
	}
	for i, opt := range options {
		switch opt := opt.(type) {
		case CalendarOption:
			if opt != nil {
				if err := opt(c); err != nil {
					return nil, err
				}
			}
		case ParseOption:
			if opt != nil {
				if err := opt(c); err != nil {
					return nil, err
				}
			}
		case TimezoneMapper:
			c.timezoneMapper = opt
		case func(string) *time.Location:
			c.timezoneMapper = TimezoneMapper(opt)
		case PropertyParser:
			if opt != nil {
				c.propertyParser = opt
			}
		case func(ContentLine) (*BaseProperty, error):
			if opt != nil {
				c.propertyParser = PropertyParser(opt)
			}
		default:
			return nil, fmt.Errorf("%w %d: %T", ErrInvalidOpArg, i, opt)
		}
	}
	return c, nil
}

func ParseCalendar(r io.Reader) (*Calendar, error) {
	// Default behavior maintains backward compatibility (strict mode)
	return ParseCalendarWithOptions(r)
}

func ParseCalendarWithOptions(r io.Reader, options ...any) (*Calendar, error) {
	state := "begin"
	c, err := NewCalendarWithOptions(options...)
	if err != nil {
		return nil, err
	}
	cs := NewCalendarStream(r)
	cont := true
	for cont {
		l, lineNo, err := cs.ReadLine()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				cont = false
			default:
				return c, err
			}
		}
		if l == nil || len(*l) == 0 {
			continue
		}
		line, err := c.propertyParser(*l)
		if err != nil {
			if errors.Is(err, ErrPropertySkipped) {
				continue
			}
			return nil, NewMalformedError(lineNo, -1, err)
		}
		if line == nil {
			continue
		}
		switch state {
		case "begin":
			switch line.IANAToken {
			case "BEGIN":
				switch line.Value {
				case "VCALENDAR":
					state = "properties"
				default:
					return nil, NewMalformedError(lineNo, -1, ErrExpectedVCalendar)
				}
			default:
				return nil, NewMalformedError(lineNo, -1, ErrExpectedBegin)
			}
		case "properties":
			switch line.IANAToken {
			case "END":
				switch line.Value {
				case "VCALENDAR":
					state = "end"
				default:
					return nil, NewMalformedError(lineNo, -1, ErrExpectedEnd)
				}
			case "BEGIN":
				state = "components"
			case string(PropertyCalscale), string(PropertyMethod), string(PropertyProductId), string(PropertyVersion), string(PropertyName), string(PropertyXWRCalName), string(PropertyXWRCalDesc), string(PropertyXWRTimezone), string(PropertyXWRCalID), string(PropertyXPublishedTTL), string(PropertyRefreshInterval), string(PropertyColor), string(PropertyDescription), string(PropertyLastModified), string(PropertyUrl), string(PropertyTzid), string(PropertyTimezoneId), string(PropertySource):
				c.CalendarProperties = append(c.CalendarProperties, CalendarProperty{*line})
			default:
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
					return nil, NewMalformedError(lineNo, -1, ErrExpectedEnd)
				}
			case "BEGIN":
				co, err := generalParseComponentWithHandler(cs, line, c.componentParseOptions()...)
				if err != nil {
					return nil, err
				}
				if co != nil {
					c.Components = append(c.Components, co)
				}
			default:
				if err := c.unknownCalendarPropertyHandler(c, state, line); err != nil {
					return nil, NewMalformedError(lineNo, -1, err)
				}
			}
		case "end":
			return nil, NewMalformedError(lineNo, -1, ErrUnexpectedCalendarEnd)
		default:
			return nil, NewMalformedError(lineNo, -1, ErrBadCalendarState)
		}
	}
	return c, nil
}

// AcceptUnknownPropertyHandler allows properties between components (non-standard but occurs in real-world ICS files)
// These properties are added to the calendar properties list
func AcceptUnknownPropertyHandler(cal *Calendar, state string, cl *BaseProperty) error {
	switch state {
	case "components":
		cal.CalendarProperties = append(cal.CalendarProperties, CalendarProperty{*cl})
		return nil
	default:
		return DefaultUnknownCalendarPropertyHandler(cal, state, cl)
	}
}

func DefaultUnknownCalendarPropertyHandler(cal *Calendar, state string, cl *BaseProperty) error {
	return ErrExpectedBeginOrEnd
}

type CalendarStream struct {
	r    io.Reader
	b    *bufio.Reader
	line int
}

func NewCalendarStream(r io.Reader) *CalendarStream {
	return &CalendarStream{
		r: r,
		b: bufio.NewReader(r),
	}
}

func (cs *CalendarStream) ReadLine() (*ContentLine, int, error) {
	r := []byte{}
	c := true
	var err error
	lineNo := cs.line + 1
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
			if errors.Is(err, io.EOF) {
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
		switch {
		case err == nil:
			if len(r) == 0 {
				c = true
			}
		case errors.Is(err, io.EOF):
			c = false
		default:
			// This must be as a result of boxing?
			if err != nil {
				err = fmt.Errorf("readline: %w", err)
			}
			return nil, lineNo, err
		}
	}
	if len(r) == 0 && err != nil {
		return nil, lineNo, fmt.Errorf("readline: %w", err)
	}
	cl := ContentLine(r)
	cs.line = lineNo
	if err != nil {
		err = fmt.Errorf("readline: %w", err)
	}
	return &cl, lineNo, err
}
