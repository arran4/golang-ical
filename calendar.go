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

// ComponentType enumerates the component names defined in RFC 5545 section 3.6.
type ComponentType string

const (
	// ComponentVCalendar is the VCALENDAR container component.
	ComponentVCalendar ComponentType = "VCALENDAR"
	// ComponentVEvent represents a VEVENT component.
	ComponentVEvent ComponentType = "VEVENT"
	// ComponentVTodo represents a VTODO component.
	ComponentVTodo ComponentType = "VTODO"
	// ComponentVJournal represents a VJOURNAL component.
	ComponentVJournal ComponentType = "VJOURNAL"
	// ComponentVFreeBusy represents a VFREEBUSY component.
	ComponentVFreeBusy ComponentType = "VFREEBUSY"
	// ComponentVTimezone represents a VTIMEZONE component.
	ComponentVTimezone ComponentType = "VTIMEZONE"
	// ComponentVAlarm represents a VALARM subcomponent.
	ComponentVAlarm ComponentType = "VALARM"
	// ComponentStandard represents a STANDARD timezone subcomponent.
	ComponentStandard ComponentType = "STANDARD"
	// ComponentDaylight represents a DAYLIGHT timezone subcomponent.
	ComponentDaylight ComponentType = "DAYLIGHT"
)

// ComponentProperty enumerates the iCalendar property names used by
// components.  Each constant is the textual property name defined in RFC 5545
// section 3.8.  These identifiers are used with methods such as
// ComponentBase.SetProperty or ComponentBase.GetProperty to manipulate
// component data without having to type property strings manually.
//
// Example (VEVENT):
//
//	cal := NewCalendar()
//	e := cal.AddEvent("20240601T120000Z-1234@example.com")
//	e.SetProperty(ComponentPropertySummary, "Sprint Review")
//	e.SetProperty(ComponentPropertyDtStart, "20240601T120000Z")
//	e.SetProperty(ComponentPropertyDtEnd, "20240601T130000Z")
//
// Using the constants ensures the property names are spelled correctly and
// allows static checking by the compiler.
type ComponentProperty Property

const (
	// ComponentPropertyUniqueId maps to the UID property (RFC 5545 section 3.8.4.7).
	// Every VEVENT or VTODO must include exactly one UID so calendar clients can
	// track the item across updates.
	// Example:
	//
	//     e := NewEvent("19960901T130000Z-123401@example.com")
	//     e.SetProperty(ComponentPropertyUniqueId, "19960901T130000Z-123401@example.com")
	ComponentPropertyUniqueId = ComponentProperty(PropertyUid) // TEXT
	// ComponentPropertyDtstamp maps to the DTSTAMP property (section 3.8.7.2).
	// DTSTAMP records when the component was created.
	// Example using the helper:
	//
	//     e.SetDtStampTime(time.Now())
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyDtstamp, time.Now().UTC().Format("20060102T150405Z"))
	ComponentPropertyDtstamp = ComponentProperty(PropertyDtstamp)
	// ComponentPropertyOrganizer maps to the ORGANIZER property (section 3.8.4.3).
	// It stores the calendar address of the meeting organizer as a CAL-ADDRESS
	// (typically "mailto:" followed by an email address).
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyOrganizer, "mailto:boss@example.com")
	//
	// Example using the helper, which automatically adds the "mailto:" prefix:
	//
	//     e.SetOrganizer("boss@example.com")
	ComponentPropertyOrganizer = ComponentProperty(PropertyOrganizer)
	// ComponentPropertyAttendee maps to the ATTENDEE property (section 3.8.4.1).
	// It lists participants invited to the event.
	// Example using the helper:
	//
	//     e.AddAttendee("dev@example.com", WithRole("REQ-PARTICIPANT"))
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyAttendee, "mailto:dev@example.com", WithRole("REQ-PARTICIPANT"))
	ComponentPropertyAttendee = ComponentProperty(PropertyAttendee)
	// ComponentPropertyAttach maps to the ATTACH property (section 3.8.1.1).
	// Example using the helper:
	//
	//     e.AddAttachmentURL("https://example.com/manual.pdf", "application/pdf")
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyAttach, "https://example.com/manual.pdf", WithFmtType("application/pdf"))
	ComponentPropertyAttach = ComponentProperty(PropertyAttach)
	// ComponentPropertyDescription maps to DESCRIPTION (section 3.8.1.5).
	// This text is presented to users as the event body or notes.
	// Example using the helper:
	//
	//     e.SetDescription("Discuss project status")
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyDescription, "Discuss project status")
	ComponentPropertyDescription = ComponentProperty(PropertyDescription) // TEXT
	// ComponentPropertyCategories maps to CATEGORIES (section 3.8.1.2).
	// Categories group similar events for filtering.
	// Example using the helper:
	//
	//     e.AddCategory("MEETING")
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyCategories, "MEETING")
	ComponentPropertyCategories = ComponentProperty(PropertyCategories) // TEXT
	// ComponentPropertyClass maps to CLASS (section 3.8.1.3).
	// CLASS controls the access level (PUBLIC, PRIVATE, CONFIDENTIAL).
	// Example using the helper:
	//
	//     e.SetClass(ClassificationPublic)
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyClass, string(ClassificationPublic))
	ComponentPropertyClass = ComponentProperty(PropertyClass) // TEXT
	// ComponentPropertyColor maps to COLOR (non-standard but common extension).
	// Example using the helper:
	//
	//     e.SetColor("#FF0000")
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyColor, "#FF0000")
	ComponentPropertyColor = ComponentProperty(PropertyColor) // TEXT
	// ComponentPropertyCreated maps to CREATED (RFC 5545 section 3.8.7.1).
	// Example using the helper:
	//     e.SetCreatedTime(time.Now())
	//
	// Example without the helper:
	//     e.SetProperty(ComponentPropertyCreated, time.Now().UTC().Format("20060102T150405Z"))
	ComponentPropertyCreated = ComponentProperty(PropertyCreated)
	// ComponentPropertySummary maps to SUMMARY (section 3.8.1.12).
	// Example using the helper:
	//
	//     e.SetSummary("Weekly Sync")
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertySummary, "Weekly Sync")
	ComponentPropertySummary = ComponentProperty(PropertySummary) // TEXT
	// ComponentPropertyDtStart maps to DTSTART (section 3.8.2.4).
	// Example using the helper:
	//
	//     e.SetStartAt(time.Now())
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyDtStart, time.Now().UTC().Format("20060102T150405Z"))
	ComponentPropertyDtStart = ComponentProperty(PropertyDtstart)
	// ComponentPropertyDtEnd maps to DTEND (section 3.8.2.2).
	// Example using the helper:
	//
	//     e.SetEndAt(time.Now().Add(1 * time.Hour))
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyDtEnd, time.Now().Add(1*time.Hour).UTC().Format("20060102T150405Z"))
	ComponentPropertyDtEnd = ComponentProperty(PropertyDtend)
	// ComponentPropertyLocation maps to LOCATION (section 3.8.1.7).
	// Example using the helper:
	//
	//     e.SetLocation("Conference Room 1")
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyLocation, "Conference Room 1")
	ComponentPropertyLocation = ComponentProperty(PropertyLocation) // TEXT
	// ComponentPropertyStatus maps to STATUS (section 3.8.1.11).
	// Example using the helper:
	//
	//     e.SetStatus(ObjectStatusConfirmed)
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyStatus, string(ObjectStatusConfirmed))
	ComponentPropertyStatus = ComponentProperty(PropertyStatus) // TEXT
	// ComponentPropertyFreebusy maps to FREEBUSY (section 3.8.2.6).
	// There is no dedicated helper for this property.
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyFreebusy, "20240601T120000Z/20240601T130000Z")
	ComponentPropertyFreebusy = ComponentProperty(PropertyFreebusy)
	// ComponentPropertyLastModified maps to LAST-MODIFIED (section 3.8.7.3).
	// Example using the helper:
	//
	//     e.SetLastModifiedAt(time.Now())
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyLastModified, time.Now().UTC().Format("20060102T150405Z"))
	ComponentPropertyLastModified = ComponentProperty(PropertyLastModified)
	// ComponentPropertyUrl maps to URL (section 3.8.4.6).
	// Example using the helper:
	//
	//     e.SetURL("https://example.com/event")
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyUrl, "https://example.com/event")
	ComponentPropertyUrl = ComponentProperty(PropertyUrl)
	// ComponentPropertyGeo maps to GEO (section 3.8.1.6).
	// Example using the helper:
	//
	//     e.SetGeo(37.386013, -122.082932)
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyGeo, "37.386013;-122.082932")
	ComponentPropertyGeo = ComponentProperty(PropertyGeo)
	// ComponentPropertyTransp maps to TRANSP (section 3.8.2.7).
	// Example using the helper:
	//
	//     e.SetTimeTransparency(TransparencyTransparent)
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyTransp, string(TransparencyTransparent))
	ComponentPropertyTransp = ComponentProperty(PropertyTransp)
	// ComponentPropertySequence maps to SEQUENCE (section 3.8.7.4).
	// Example using the helper:
	//
	//     e.SetSequence(2)
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertySequence, "2")
	ComponentPropertySequence = ComponentProperty(PropertySequence)
	// ComponentPropertyExdate maps to EXDATE (section 3.8.5.1).
	// Example using the helper:
	//
	//     e.AddExdate("20240608T120000Z")
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyExdate, "20240608T120000Z")
	ComponentPropertyExdate = ComponentProperty(PropertyExdate)
	// ComponentPropertyExrule represents the deprecated EXRULE property
	// (originally RFC 2445 section 4.8.5.2).
	// Example using the helper:
	//
	//     e.AddExrule("FREQ=DAILY")
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyExrule, "FREQ=DAILY")
	ComponentPropertyExrule = ComponentProperty(PropertyExrule)
	// ComponentPropertyRdate maps to RDATE (section 3.8.5.2).
	// Example using the helper:
	//
	//     e.AddRdate("20240615T120000Z")
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyRdate, "20240615T120000Z")
	ComponentPropertyRdate = ComponentProperty(PropertyRdate)
	// ComponentPropertyRrule maps to RRULE (section 3.8.5.3).
	// Example using the helper:
	//
	//     e.AddRrule("FREQ=WEEKLY;BYDAY=MO")
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyRrule, "FREQ=WEEKLY;BYDAY=MO")
	ComponentPropertyRrule = ComponentProperty(PropertyRrule)
	// ComponentPropertyAction maps to the ACTION property (section 3.8.6.1).
	// Example using the helper:
	//
	//     alarm.SetAction(ActionDisplay)
	//
	// Example without the helper:
	//
	//     alarm.SetProperty(ComponentPropertyAction, "DISPLAY")
	ComponentPropertyAction = ComponentProperty(PropertyAction)
	// ComponentPropertyTrigger maps to TRIGGER (section 3.8.6.3).
	// Example using the helper:
	//
	//     alarm.SetTrigger("-PT15M")
	//
	// Example without the helper:
	//
	//     alarm.SetProperty(ComponentPropertyTrigger, "-PT15M")
	ComponentPropertyTrigger = ComponentProperty(PropertyTrigger)
	// ComponentPropertyPriority maps to PRIORITY (section 3.8.1.9).
	// Example using the helper:
	//
	//     e.SetPriority(1)
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyPriority, "1")
	ComponentPropertyPriority = ComponentProperty(PropertyPriority)
	// ComponentPropertyResources maps to RESOURCES (section 3.8.1.10).
	// Example using the helper:
	//
	//     e.SetResources("PROJECTOR")
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyResources, "PROJECTOR")
	ComponentPropertyResources = ComponentProperty(PropertyResources)
	// ComponentPropertyCompleted maps to COMPLETED (section 3.8.2.1).
	// Example using the helper:
	//
	//     todo.SetCompletedAt(time.Date(2024, 6, 1, 17, 0, 0, 0, time.UTC))
	//
	// Example without the helper:
	//
	//     todo.SetProperty(ComponentPropertyCompleted, time.Date(2024, 6, 1, 17, 0, 0, 0, time.UTC).UTC().Format("20060102T150405Z"))
	ComponentPropertyCompleted = ComponentProperty(PropertyCompleted)
	// ComponentPropertyDue maps to DUE (section 3.8.2.3).
	// Example using the helper:
	//
	//     todo.SetDueAt(time.Date(2024, 6, 7, 12, 0, 0, 0, time.UTC))
	//
	// Example without the helper:
	//
	//     todo.SetProperty(ComponentPropertyDue, time.Date(2024, 6, 7, 12, 0, 0, 0, time.UTC).UTC().Format("20060102T150405Z"))
	ComponentPropertyDue = ComponentProperty(PropertyDue)
	// ComponentPropertyPercentComplete maps to PERCENT-COMPLETE (section 3.8.1.8).
	// Example using the helper:
	//
	//     todo.SetPercentComplete(50)
	//
	// Example without the helper:
	//
	//     todo.SetProperty(ComponentPropertyPercentComplete, "50")
	ComponentPropertyPercentComplete = ComponentProperty(PropertyPercentComplete)
	// ComponentPropertyTzid maps to TZID (section 3.8.3.1).
	// There is no dedicated helper for this property.
	// Example without the helper:
	//
	//     timezone.SetProperty(ComponentPropertyTzid, "America/New_York")
	ComponentPropertyTzid = ComponentProperty(PropertyTzid)
	// ComponentPropertyComment maps to COMMENT (section 3.8.1.4).
	// Example using the helper:
	//
	//     e.AddComment("Bring slides")
	//
	// Example without the helper:
	//
	//     e.AddProperty(ComponentPropertyComment, "Bring slides")
	ComponentPropertyComment = ComponentProperty(PropertyComment)
	// ComponentPropertyRelatedTo maps to RELATED-TO (section 3.8.4.5).
	// There is no dedicated helper for this property.
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyRelatedTo, "19960901T130000Z-123401@example.com")
	ComponentPropertyRelatedTo = ComponentProperty(PropertyRelatedTo)
	// ComponentPropertyMethod maps to METHOD (section 3.7.2).
	// Example using the helper:
	//
	//     cal.SetMethod(MethodPublish)
	//
	// Example without the helper:
	//
	//     cal.SetProperty(PropertyMethod, string(MethodPublish))
	ComponentPropertyMethod = ComponentProperty(PropertyMethod)
	// ComponentPropertyRecurrenceId maps to RECURRENCE-ID (section 3.8.4.4).
	// There is no dedicated helper for this property.
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyRecurrenceId, "20240608T120000Z")
	ComponentPropertyRecurrenceId = ComponentProperty(PropertyRecurrenceId)
	// ComponentPropertyDuration maps to DURATION (section 3.8.2.5).
	// Example using the helper:
	//
	//     e.SetDuration(time.Hour)
	//
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyDuration, "PT1H")
	ComponentPropertyDuration = ComponentProperty(PropertyDuration)
	// ComponentPropertyContact maps to CONTACT (section 3.8.4.2).
	// There is no dedicated helper for this property.
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyContact, "mailto:hr@example.com")
	ComponentPropertyContact = ComponentProperty(PropertyContact)
	// ComponentPropertyRequestStatus maps to REQUEST-STATUS (section 3.8.8.3).
	// There is no dedicated helper for this property.
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyRequestStatus, "2.0;Success")
	ComponentPropertyRequestStatus = ComponentProperty(PropertyRequestStatus)
	// ComponentPropertyRDate is kept for backward compatibility and is
	// equivalent to ComponentPropertyRdate.
	// There is no helper for this property.
	// Example without the helper:
	//
	//     e.SetProperty(ComponentPropertyRDate, "20240615T120000Z")
	ComponentPropertyRDate = ComponentProperty(PropertyRdate)
)

// Required reports whether the property is mandatory for the given component
// type according to the tables in RFC 5545 section 3.6.  The implementation
// currently handles a subset of properties based on common usage.
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

// Exclusive returns the set of properties that cannot coexist with cp on the
// provided component as described in RFC 5545 section 3.6.  Only a limited set
// of relationships are implemented.
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

// Singular reports whether cp is restricted to a single occurrence within the
// given component according to RFC 5545 section 3.6.  Only common properties are
// recognized.
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

// Optional reports whether cp may appear in the component per the property
// tables in RFC 5545 section 3.6.  This helper only covers selected properties.
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

// Multiple reports whether cp may occur more than once within the component as
// described in RFC 5545 section 3.6.  The coverage is not exhaustive.
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

// Property enumerates iCalendar property names as defined primarily in RFC 5545
// section 3.8.  Each constant maps to its textual representation.
const (
	// PropertyCalscale corresponds to CALSCALE (section 3.7.1).
	// Example using the helper:
	//
	//     cal := NewCalendar()
	//     cal.SetCalscale("GREGORIAN")
	//
	// Example without the helper:
	//
	//     cal.SetProperty(PropertyCalscale, "GREGORIAN")
	PropertyCalscale Property = "CALSCALE" // TEXT
	// PropertyMethod corresponds to METHOD (section 3.7.2).
	// Example using the helper:
	//
	//     cal := NewCalendar()
	//     cal.SetMethod(MethodPublish)
	//
	// Example without the helper:
	//
	//     cal.SetProperty(PropertyMethod, string(MethodPublish))
	PropertyMethod Property = "METHOD" // TEXT
	// PropertyProductId corresponds to PRODID (section 3.7.3).
	// Example using the helper:
	//
	//     cal := NewCalendar()
	//     cal.SetProductId("-//Example Corp//Calendar")
	//
	// Example without the helper:
	//
	//     cal.SetProperty(PropertyProductId, "-//Example Corp//Calendar")
	PropertyProductId Property = "PRODID" // TEXT
	// PropertyVersion corresponds to VERSION (section 3.7.4).
	// Example using the helper:
	//
	//     cal := NewCalendar()
	//     cal.SetVersion("2.0")
	//
	// Example without the helper:
	//
	//     cal.SetProperty(PropertyVersion, "2.0")
	PropertyVersion Property = "VERSION" // TEXT
	// PropertyXPublishedTTL is a common extension used to signal how long
	// the calendar data may be cached.
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetXPublishedTTL("PT1H")
	//
	PropertyXPublishedTTL Property = "X-PUBLISHED-TTL"
	// PropertyRefreshInterval indicates how often clients should refresh the
	// calendar (section 3.7.4.3 of RFC 7986).
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetRefreshInterval("PT1H")
	//
	PropertyRefreshInterval Property = "REFRESH-INTERVAL;VALUE=DURATION"
	// PropertyAttach adds a binary or URI attachment (section 3.8.1.1).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddAttachment("https://example.com/info.pdf")
	//
	PropertyAttach Property = "ATTACH"
	// PropertyCategories corresponds to CATEGORIES (section 3.8.1.2).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddCategory("MEETING")
	//
	PropertyCategories Property = "CATEGORIES" // TEXT
	// PropertyClass corresponds to CLASS (section 3.8.1.3).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetClass(ClassPublic)
	//
	PropertyClass Property = "CLASS" // TEXT
	// PropertyColor is a common extension for calendar color.
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetColor("#00AA00")
	//
	PropertyColor Property = "COLOR" // TEXT
	// PropertyComment corresponds to COMMENT (section 3.8.1.4).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddComment("Bring projector")
	//
	PropertyComment Property = "COMMENT" // TEXT
	// PropertyDescription corresponds to DESCRIPTION (section 3.8.1.5).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetDescription("Sprint review meeting")
	//
	PropertyDescription Property = "DESCRIPTION" // TEXT
	// PropertyXWRCalDesc is an Apple extension describing the calendar.
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetXWRCalDesc("Team schedule")
	//
	PropertyXWRCalDesc Property = "X-WR-CALDESC"
	// PropertyGeo stores geographic position in "lat;lon" format
	// (section 3.8.1.6).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetGeo(51.5, -0.1)
	//
	PropertyGeo Property = "GEO"
	// PropertyLocation corresponds to LOCATION (section 3.8.1.7).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetLocation("Conference Room")
	//
	PropertyLocation Property = "LOCATION" // TEXT
	// PropertyPercentComplete indicates task completion percentage
	// (section 3.8.1.8).
	// Example:
	//
	//     todo := NewTodo("id")
	//     todo.SetPercentComplete(50)
	//
	PropertyPercentComplete Property = "PERCENT-COMPLETE"
	// PropertyPriority sets the task or event priority (section 3.8.1.9).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetPriority(1)
	//
	PropertyPriority Property = "PRIORITY"
	// PropertyResources lists resources needed (section 3.8.1.10).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetResources("PROJECTOR")
	//
	PropertyResources Property = "RESOURCES" // TEXT
	// PropertyStatus sets the overall status (section 3.8.1.11).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetStatus(ObjectStatusConfirmed)
	//
	PropertyStatus Property = "STATUS" // TEXT
	// PropertySummary holds the title of the component (section 3.8.1.12).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetSummary("Weekly Sync")
	//
	PropertySummary Property = "SUMMARY" // TEXT
	// PropertyCompleted records when a VTODO was completed (section 3.8.2.1).
	// Example:
	//
	//     todo := NewTodo("id")
	//     todo.SetCompletedAt(time.Now())
	//
	PropertyCompleted Property = "COMPLETED"
	// PropertyDtend gives the end time of a VEVENT (section 3.8.2.2).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetEndAt(time.Now().Add(1*time.Hour))
	//
	PropertyDtend Property = "DTEND"
	// PropertyDue sets the due date of a VTODO (section 3.8.2.3).
	// Example:
	//
	//     todo := NewTodo("id")
	//     todo.SetDueAt(time.Now().Add(24 * time.Hour))
	//
	PropertyDue Property = "DUE"
	// PropertyDtstart defines the start time of the component (section 3.8.2.4).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetStartAt(time.Now())
	//
	PropertyDtstart Property = "DTSTART"
	// PropertyDuration specifies the duration of the event (section 3.8.2.5).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetProperty(ComponentPropertyDuration, "PT1H")
	//
	PropertyDuration Property = "DURATION"
	// PropertyFreebusy conveys free/busy time information (section 3.8.2.6).
	// Example:
	//
	//     fb := NewFreeBusy("id")
	//     fb.SetProperty(ComponentPropertyFreebusy, "20240601T120000Z/20240601T130000Z")
	//
	PropertyFreebusy Property = "FREEBUSY"
	// PropertyTransp corresponds to TRANSP (section 3.8.2.7).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetTimeTransparency(Transparent)
	//
	PropertyTransp Property = "TRANSP" // TEXT
	// PropertyTzid identifies the timezone of a VTIMEZONE (section 3.8.3.1).
	// Example:
	//
	//     tz := NewTimezone("America/New_York")
	//     tz.SetProperty(ComponentPropertyTzid, "America/New_York")
	//
	PropertyTzid Property = "TZID" // TEXT
	// PropertyTzname gives the customary name for a timezone (section 3.8.3.2).
	// Example:
	//
	//     tz := NewTimezone("America/New_York")
	//     tz.SetProperty(ComponentPropertyTzname, "EST")
	//
	PropertyTzname Property = "TZNAME" // TEXT
	// PropertyTzoffsetfrom specifies the offset before a transition
	// (section 3.8.3.3).
	PropertyTzoffsetfrom Property = "TZOFFSETFROM"
	// PropertyTzoffsetto specifies the offset after a transition
	// (section 3.8.3.4).
	PropertyTzoffsetto Property = "TZOFFSETTO"
	// PropertyTzurl points to timezone information (section 3.8.3.5).
	// Example:
	//
	//     tz := NewTimezone("America/New_York")
	//     tz.SetProperty(ComponentPropertyTzurl, "https://tz.example.com/nyc")
	//
	PropertyTzurl Property = "TZURL"
	// PropertyAttendee lists a participant (section 3.8.4.1).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddAttendee("mailto:dev@example.com")
	//
	PropertyAttendee Property = "ATTENDEE"
	// PropertyContact supplies contact information (section 3.8.4.2).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetProperty(ComponentPropertyContact, "mailto:hr@example.com")
	//
	PropertyContact Property = "CONTACT" // TEXT
	// PropertyOrganizer gives the organizer's address (section 3.8.4.3).
	// Example without helper:
	//
	//     e := NewEvent("id")
	//     e.SetProperty(ComponentPropertyOrganizer, "mailto:boss@example.com")
	//
	// The SetOrganizer helper automatically prefixes the value with "mailto:"
	// when needed:
	//
	//     e.SetOrganizer("boss@example.com")
	PropertyOrganizer Property = "ORGANIZER"
	// PropertyRecurrenceId identifies a specific recurrence (section 3.8.4.4).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetProperty(ComponentPropertyRecurrenceId, "20240608T120000Z")
	//
	PropertyRecurrenceId Property = "RECURRENCE-ID"
	// PropertyRelatedTo corresponds to RELATED-TO (section 3.8.4.5).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetProperty(ComponentPropertyRelatedTo, "19960901T130000Z-123401@example.com")
	//
	PropertyRelatedTo Property = "RELATED-TO" // TEXT
	// PropertyUrl provides a link to additional information (section 3.8.4.6).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetURL("https://example.com/event")
	//
	PropertyUrl Property = "URL"
	// PropertyUid holds the globally unique identifier (section 3.8.4.7).
	// Example:
	//
	//     e := NewEvent("19960901T130000Z-123401@example.com")
	//     e.SetProperty(ComponentPropertyUniqueId, "19960901T130000Z-123401@example.com")
	//
	PropertyUid Property = "UID" // TEXT
	// PropertyExdate excludes a recurrence date (section 3.8.5.1).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddExdate("20240608T120000Z")
	//
	PropertyExdate Property = "EXDATE"
	// PropertyExrule is deprecated but represents exception rules for
	// recurrence (RFC 2445 section 4.8.5.2).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddExrule("FREQ=WEEKLY;BYDAY=MO")
	//
	PropertyExrule Property = "EXRULE"
	// PropertyRdate specifies additional recurrence dates (section 3.8.5.2).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddRdate("20240615T120000Z")
	//
	PropertyRdate Property = "RDATE"
	// PropertyRrule defines a recurrence rule (section 3.8.5.3).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.AddRrule("FREQ=DAILY")
	//
	PropertyRrule Property = "RRULE"
	// PropertyAction corresponds to ACTION (section 3.8.6.1).
	// Example:
	//
	//     alarm := NewAlarm()
	//     alarm.SetProperty(ComponentPropertyAction, "DISPLAY")
	//
	PropertyAction Property = "ACTION" // TEXT
	// PropertyRepeat indicates how often to repeat an alarm (section 3.8.6.2).
	PropertyRepeat Property = "REPEAT"
	// PropertyTrigger defines when an alarm triggers (section 3.8.6.3).
	// Example:
	//
	//     alarm := NewAlarm()
	//     alarm.SetProperty(ComponentPropertyTrigger, "-PT15M")
	//
	PropertyTrigger Property = "TRIGGER"
	// PropertyCreated records the creation time (section 3.8.7.1).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetCreatedTime(time.Now())
	//
	PropertyCreated Property = "CREATED"
	// PropertyDtstamp is the creation timestamp (section 3.8.7.2).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetDtStampTime(time.Now())
	//
	PropertyDtstamp Property = "DTSTAMP"
	// PropertyLastModified records the last modification time (section 3.8.7.3).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetModifiedAt(time.Now())
	//
	PropertyLastModified Property = "LAST-MODIFIED"
	// PropertyRequestStatus conveys the status of a scheduling request
	// (section 3.8.8.3).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetProperty(ComponentPropertyRequestStatus, "2.0;Success")
	//
	PropertyRequestStatus Property = "REQUEST-STATUS" // TEXT
	// PropertyName is an extension naming the calendar.
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetName("Company Events")
	//
	PropertyName Property = "NAME"
	// PropertyXWRCalName stores the display name for Apple clients.
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetXWRCalName("Company Events")
	//
	PropertyXWRCalName Property = "X-WR-CALNAME"
	// PropertyXWRTimezone defines the default timezone for the calendar.
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetXWRTimezone("America/New_York")
	//
	PropertyXWRTimezone Property = "X-WR-TIMEZONE"
	// PropertySequence increments on each update to an item (section 3.8.7.4).
	// Example:
	//
	//     e := NewEvent("id")
	//     e.SetSequence(2)
	//
	PropertySequence Property = "SEQUENCE"
	// PropertyXWRCalID is an Apple extension storing a stable calendar ID.
	// Example:
	//
	//     cal := NewCalendar()
	//     cal.SetXWRCalID("a1b2c3")
	//
	PropertyXWRCalID Property = "X-WR-RELCALID"
	// PropertyTimezoneId is defined in RFC 9074 for naming embedded
	// VTIMEZONE components.
	// Example:
	//
	//     tz := NewTimezone("America/New_York")
	//     tz.SetTimezoneId("America/New_York")
	//
	PropertyTimezoneId Property = "TIMEZONE-ID"
)

type Parameter string

// IsQuoted reports whether the parameter's value should be quoted when serialized.
// RFC 5545 section 3.2 specifies ALTREP as the only standard parameter requiring quotes.
func (p Parameter) IsQuoted() bool {
	switch p {
	case ParameterAltrep:
		return true
	}
	return false
}

const (
	// ParameterAltrep references an alternate text representation (section 3.2.1).
	ParameterAltrep Parameter = "ALTREP"
	// ParameterCn provides a common name (section 3.2.2).
	ParameterCn Parameter = "CN"
	// ParameterCutype defines the calendar user type (section 3.2.3).
	ParameterCutype Parameter = "CUTYPE"
	// ParameterDelegatedFrom lists participants the request was delegated from (section 3.2.4).
	ParameterDelegatedFrom Parameter = "DELEGATED-FROM"
	// ParameterDelegatedTo lists participants the request was delegated to (section 3.2.5).
	ParameterDelegatedTo Parameter = "DELEGATED-TO"
	// ParameterDir gives a reference to directory information (section 3.2.6).
	ParameterDir Parameter = "DIR"
	// ParameterEncoding defines inline attachment encoding (section 3.2.7).
	ParameterEncoding Parameter = "ENCODING"
	// ParameterFmttype is the content type for a binary attachment (section 3.2.8).
	ParameterFmttype Parameter = "FMTTYPE"
	// ParameterFbtype specifies free/busy time type (section 3.2.9).
	ParameterFbtype Parameter = "FBTYPE"
	// ParameterLanguage indicates the language for text values (section 3.2.10).
	ParameterLanguage Parameter = "LANGUAGE"
	// ParameterMember identifies group membership (section 3.2.11).
	ParameterMember Parameter = "MEMBER"
	// ParameterParticipationStatus holds participation status (section 3.2.12).
	ParameterParticipationStatus Parameter = "PARTSTAT"
	// ParameterRange is used with RECURRENCE-ID (section 3.2.13).
	ParameterRange Parameter = "RANGE"
	// ParameterRelated indicates the relationship type for FREEBUSY (section 3.2.14).
	ParameterRelated Parameter = "RELATED"
	// ParameterReltype specifies relationship type for RELATED-TO (section 3.2.15).
	ParameterReltype Parameter = "RELTYPE"
	// ParameterRole indicates participant role (section 3.2.16).
	ParameterRole Parameter = "ROLE"
	// ParameterRsvp indicates whether a response is requested (section 3.2.17).
	ParameterRsvp Parameter = "RSVP"
	// ParameterSentBy gives the address responsible for sending a request (section 3.2.18).
	ParameterSentBy Parameter = "SENT-BY"
	// ParameterTzid references a time zone identifier (section 3.2.19).
	ParameterTzid Parameter = "TZID"
	// ParameterValue sets the value data type of the property (section 3.2.20).
	ParameterValue Parameter = "VALUE"
)

type ValueDataType string

// ValueDataType lists the VALUE parameter types described in RFC 5545 section 3.3.
const (
	// ValueDataTypeBinary represents binary data (section 3.3.1).
	ValueDataTypeBinary ValueDataType = "BINARY"
	// ValueDataTypeBoolean represents boolean values (section 3.3.2).
	ValueDataTypeBoolean ValueDataType = "BOOLEAN"
	// ValueDataTypeCalAddress represents a calendar address (section 3.3.3).
	ValueDataTypeCalAddress ValueDataType = "CAL-ADDRESS"
	// ValueDataTypeDate represents a DATE value (section 3.3.4).
	ValueDataTypeDate ValueDataType = "DATE"
	// ValueDataTypeDateTime represents a DATE-TIME (section 3.3.5).
	ValueDataTypeDateTime ValueDataType = "DATE-TIME"
	// ValueDataTypeDuration represents a DURATION (section 3.3.6).
	ValueDataTypeDuration ValueDataType = "DURATION"
	// ValueDataTypeFloat represents floating point values (section 3.3.7).
	ValueDataTypeFloat ValueDataType = "FLOAT"
	// ValueDataTypeInteger represents integer values (section 3.3.8).
	ValueDataTypeInteger ValueDataType = "INTEGER"
	// ValueDataTypePeriod represents a PERIOD value (section 3.3.9).
	ValueDataTypePeriod ValueDataType = "PERIOD"
	// ValueDataTypeRecur represents a RECUR value (section 3.3.10).
	ValueDataTypeRecur ValueDataType = "RECUR"
	// ValueDataTypeText represents a TEXT value (section 3.3.11).
	ValueDataTypeText ValueDataType = "TEXT"
	// ValueDataTypeTime represents a TIME value (section 3.3.12).
	ValueDataTypeTime ValueDataType = "TIME"
	// ValueDataTypeUri represents a URI (section 3.3.13).
	ValueDataTypeUri ValueDataType = "URI"
	// ValueDataTypeUtcOffset represents UTC-OFFSET (section 3.3.14).
	ValueDataTypeUtcOffset ValueDataType = "UTC-OFFSET"
)

type CalendarUserType string

// CalendarUserType enumerates the CUTYPE parameter values from RFC 5545 section 3.2.3.
const (
	// CalendarUserTypeIndividual identifies an individual calendar user.
	CalendarUserTypeIndividual CalendarUserType = "INDIVIDUAL"
	// CalendarUserTypeGroup identifies a group of users.
	CalendarUserTypeGroup CalendarUserType = "GROUP"
	// CalendarUserTypeResource identifies a physical resource.
	CalendarUserTypeResource CalendarUserType = "RESOURCE"
	// CalendarUserTypeRoom identifies a room resource.
	CalendarUserTypeRoom CalendarUserType = "ROOM"
	// CalendarUserTypeUnknown is used when the user type is unknown.
	CalendarUserTypeUnknown CalendarUserType = "UNKNOWN"
)

func (cut CalendarUserType) KeyValue(_ ...interface{}) (string, []string) {
	return string(ParameterCutype), []string{string(cut)}
}

type FreeBusyTimeType string

// FreeBusyTimeType enumerates the FBTYPE parameter values used with FREEBUSY
// properties (RFC 5545 section 3.2.9).
const (
	// FreeBusyTimeTypeFree indicates the time is free.
	FreeBusyTimeTypeFree FreeBusyTimeType = "FREE"
	// FreeBusyTimeTypeBusy indicates the time is busy.
	FreeBusyTimeTypeBusy FreeBusyTimeType = "BUSY"
	// FreeBusyTimeTypeBusyUnavailable indicates the time is busy and unavailable.
	FreeBusyTimeTypeBusyUnavailable FreeBusyTimeType = "BUSY-UNAVAILABLE"
	// FreeBusyTimeTypeBusyTentative indicates tentative busy time.
	FreeBusyTimeTypeBusyTentative FreeBusyTimeType = "BUSY-TENTATIVE"
)

type ParticipationStatus string

// ParticipationStatus enumerates the PARTSTAT parameter values from RFC 5545 section 3.2.12.
const (
	// ParticipationStatusNeedsAction indicates a pending reply.
	ParticipationStatusNeedsAction ParticipationStatus = "NEEDS-ACTION"
	// ParticipationStatusAccepted indicates acceptance.
	ParticipationStatusAccepted ParticipationStatus = "ACCEPTED"
	// ParticipationStatusDeclined indicates the invitation was declined.
	ParticipationStatusDeclined ParticipationStatus = "DECLINED"
	// ParticipationStatusTentative indicates a tentative reply.
	ParticipationStatusTentative ParticipationStatus = "TENTATIVE"
	// ParticipationStatusDelegated indicates delegation to another party.
	ParticipationStatusDelegated ParticipationStatus = "DELEGATED"
	// ParticipationStatusCompleted indicates the task has been completed.
	ParticipationStatusCompleted ParticipationStatus = "COMPLETED"
	// ParticipationStatusInProcess indicates work is in progress.
	ParticipationStatusInProcess ParticipationStatus = "IN-PROCESS"
)

func (ps ParticipationStatus) KeyValue(_ ...interface{}) (string, []string) {
	return string(ParameterParticipationStatus), []string{string(ps)}
}

type ObjectStatus string

// ObjectStatus enumerates allowed STATUS property values for calendar objects
// (RFC 5545 section 3.8.1.11).
const (
	// ObjectStatusTentative indicates the object is tentative.
	ObjectStatusTentative ObjectStatus = "TENTATIVE"
	// ObjectStatusConfirmed indicates the object is confirmed.
	ObjectStatusConfirmed ObjectStatus = "CONFIRMED"
	// ObjectStatusCancelled indicates the object is cancelled.
	ObjectStatusCancelled ObjectStatus = "CANCELLED"
	// ObjectStatusNeedsAction indicates further action is required.
	ObjectStatusNeedsAction ObjectStatus = "NEEDS-ACTION"
	// ObjectStatusCompleted indicates completion.
	ObjectStatusCompleted ObjectStatus = "COMPLETED"
	// ObjectStatusInProcess indicates processing is ongoing.
	ObjectStatusInProcess ObjectStatus = "IN-PROCESS"
	// ObjectStatusDraft indicates a draft state.
	ObjectStatusDraft ObjectStatus = "DRAFT"
	// ObjectStatusFinal indicates a final state.
	ObjectStatusFinal ObjectStatus = "FINAL"
)

func (ps ObjectStatus) KeyValue(_ ...interface{}) (string, []string) {
	return string(PropertyStatus), []string{string(ps)}
}

type RelationshipType string

// RelationshipType enumerates RELTYPE parameter values for RELATED-TO
// properties (RFC 5545 section 3.2.15).
const (
	// RelationshipTypeChild indicates a child relationship.
	RelationshipTypeChild RelationshipType = "CHILD"
	// RelationshipTypeParent indicates a parent relationship.
	RelationshipTypeParent RelationshipType = "PARENT"
	// RelationshipTypeSibling indicates a sibling relationship.
	RelationshipTypeSibling RelationshipType = "SIBLING"
)

type ParticipationRole string

// ParticipationRole enumerates the ROLE parameter values for participants
// (RFC 5545 section 3.2.16).
const (
	// ParticipationRoleChair designates the chair of the meeting.
	ParticipationRoleChair ParticipationRole = "CHAIR"
	// ParticipationRoleReqParticipant indicates a required participant.
	ParticipationRoleReqParticipant ParticipationRole = "REQ-PARTICIPANT"
	// ParticipationRoleOptParticipant indicates an optional participant.
	ParticipationRoleOptParticipant ParticipationRole = "OPT-PARTICIPANT"
	// ParticipationRoleNonParticipant indicates a non-participant observer.
	ParticipationRoleNonParticipant ParticipationRole = "NON-PARTICIPANT"
)

func (pr ParticipationRole) KeyValue(_ ...interface{}) (string, []string) {
	return string(ParameterRole), []string{string(pr)}
}

type Action string

// Action enumerates VALARM ACTION property values (RFC 5545 section 3.8.6.1).
const (
	// ActionAudio plays an audio alert.
	ActionAudio Action = "AUDIO"
	// ActionDisplay shows display text.
	ActionDisplay Action = "DISPLAY"
	// ActionEmail sends an email message.
	ActionEmail Action = "EMAIL"
	// ActionProcedure invokes a procedure.
	ActionProcedure Action = "PROCEDURE"
)

type Classification string

// Classification enumerates CLASS property values (RFC 5545 section 3.8.1.3).
const (
	// ClassificationPublic marks information as public.
	ClassificationPublic Classification = "PUBLIC"
	// ClassificationPrivate marks information as private.
	ClassificationPrivate Classification = "PRIVATE"
	// ClassificationConfidential marks information as confidential.
	ClassificationConfidential Classification = "CONFIDENTIAL"
)

type Method string

// Method enumerates METHOD property values used with scheduling messages
// (RFC 5545 section 3.7.2).
const (
	// MethodPublish publishes a calendar.
	MethodPublish Method = "PUBLISH"
	// MethodRequest requests scheduling.
	MethodRequest Method = "REQUEST"
	// MethodReply sends a scheduling reply.
	MethodReply Method = "REPLY"
	// MethodAdd adds additional information.
	MethodAdd Method = "ADD"
	// MethodCancel cancels a previously scheduled object.
	MethodCancel Method = "CANCEL"
	// MethodRefresh requests a resend of a calendar.
	MethodRefresh Method = "REFRESH"
	// MethodCounter sends a counter proposal.
	MethodCounter Method = "COUNTER"
	// MethodDeclinecounter declines a counter proposal.
	MethodDeclinecounter Method = "DECLINECOUNTER"
)

type CalendarProperty struct {
	BaseProperty
}

// Calendar represents a VCALENDAR object.  RFC 5545 section 3.6 says:
// "A 'VCALENDAR' object MUST include the 'PRODID' and 'VERSION' properties" and
// it must contain at least one component such as VEVENT.  NewCalendar and
// NewCalendarFor create a calendar populated with those required fields.
type Calendar struct {
	Components         []Component
	CalendarProperties []CalendarProperty
}

// NewCalendar returns a basic Calendar using a default product identifier.
// The returned calendar satisfies the minimum requirements of RFC 5545 by
// including the VERSION and PRODID properties.
func NewCalendar() *Calendar {
	return NewCalendarFor("arran4")
}

// NewCalendarFor constructs a Calendar for the given service.  The VERSION
// property is set to "2.0" as defined in RFC 5545 section 3.7.4 and PRODID is
// populated using the provided service identifier per section 3.7.3.
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

// SerializationConfiguration controls how calendars and components are written
// out.  MaxLength and PropertyMaxLength correspond to the 75 octet line length
// recommendations from RFC 5545 section 3.1.  NewLine selects the line
// termination sequence.
type SerializationConfiguration struct {
	MaxLength         int
	NewLine           string
	PropertyMaxLength int
}

// parseSerializeOps interprets the optional arguments provided to Serialize or
// SerializeTo.  It accepts WithLineLength, WithNewLine or a
// *SerializationConfiguration.  Unsupported types return an error.
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

// defaultSerializationOptions returns the default values used for calendar
// serialization.  The line length defaults to 75 characters as recommended by
// the specification and the newline is platform specific.
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

// ParseCalendarFromUrl retrieves an iCalendar object from the provided URL and
// parses it.  Many calendaring services expose feeds over HTTP.  This helper
// performs the request and then calls ParseCalendar on the response body.  The
// resulting Calendar adheres to the grammar in RFC 5545.
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

// parseCalendarFromHttpRequest executes the HTTP request using the supplied
// client and parses the response body.  It is a helper for
// ParseCalendarFromUrl and allows custom HTTP clients to be injected for
// testing or advanced configuration.
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

// ParseCalendar reads a VCALENDAR object from r.  It implements the grammar
// described in RFC 5545 section 3.4 which states:
//
//	"The iCalendar object MUST begin with the BEGIN property with a value of
//	 VCALENDAR and end with the END property with a value of VCALENDAR."
//
// Lines between those markers are parsed into properties and components.
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
			default:
				// Unknown property names are retained to ensure
				// that vendor extensions or future RFC updates
				// are not lost when the calendar is parsed and
				// serialized again.
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

// CalendarStream reads content lines from an iCalendar stream.  The reader
// handles line folding as described in RFC 5545 section 3.1 so that callers see
// logical lines without CRLF continuations.  Lines in an iCalendar file are
// "folded" by inserting CRLF followed by a single whitespace.  This type hides
// that detail by returning unfolded lines.
type CalendarStream struct {
	r io.Reader
	b *bufio.Reader
}

// NewCalendarStream wraps r so the caller can read unfolded content lines.  The
// underlying reader is buffered, and each call to ReadLine returns a single
// logical line without trailing CRLF.  See RFC 5545 section 3.1 for details on
// the iCalendar line folding mechanism.
func NewCalendarStream(r io.Reader) *CalendarStream {
	return &CalendarStream{
		r: r,
		b: bufio.NewReader(r),
	}
}

// ReadLine reads the next unfolded content line from the stream.  Folding is
// processed per RFC 5545 section 3.1 where any CRLF followed by a space or
// horizontal tab is removed.  The returned ContentLine does not include the
// terminating newline sequence.
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
