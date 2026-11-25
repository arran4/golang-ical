package microsoft

import (
	ical "github.com/arran4/golang-ical"
)

const (
	// PropertyCalend is the X-CALEND property
	PropertyCalend ical.Property = "X-CALEND"
	// PropertyCalstart is the X-CALSTART property
	PropertyCalstart ical.Property = "X-CALSTART"
	// PropertyClipend is the X-CLIPEND property
	PropertyClipend ical.Property = "X-CLIPEND"
	// PropertyClipstart is the X-CLIPSTART property
	PropertyClipstart ical.Property = "X-CLIPSTART"
	// PropertyMicrosoftCalscale is the X-MICROSOFT-CALSCALE property
	PropertyMicrosoftCalscale ical.Property = "X-MICROSOFT-CALSCALE"
	// PropertyMsOlkForceinspectoropen is the X-MS-OLK-FORCEINSPECTOROPEN property
	PropertyMsOlkForceinspectoropen ical.Property = "X-MS-OLK-FORCEINSPECTOROPEN"
	// PropertyMsWkhrdays is the X-MS-WKHRDAYS property
	PropertyMsWkhrdays ical.Property = "X-MS-WKHRDAYS"
	// PropertyMsWkhrend is the X-MS-WKHREND property
	PropertyMsWkhrend ical.Property = "X-MS-WKHREND"
	// PropertyMsWkhrstart is the X-MS-WKHRSTART property
	PropertyMsWkhrstart ical.Property = "X-MS-WKHRSTART"
	// PropertyOwner is the X-OWNER property
	PropertyOwner ical.Property = "X-OWNER"
	// PropertyPrimaryCalendar is the X-PRIMARY-CALENDAR property
	PropertyPrimaryCalendar ical.Property = "X-PRIMARY-CALENDAR"
)

const (
	// ComponentPropertyAltDesc is the X-ALT-DESC component property
	ComponentPropertyAltDesc ical.ComponentProperty = "X-ALT-DESC"
	// ComponentPropertyMicrosoftCdoAlldayevent is the X-MICROSOFT-CDO-ALLDAYEVENT component property
	ComponentPropertyMicrosoftCdoAlldayevent ical.ComponentProperty = "X-MICROSOFT-CDO-ALLDAYEVENT"
	// ComponentPropertyMicrosoftCdoApptSequence is the X-MICROSOFT-CDO-APPT-SEQUENCE component property
	ComponentPropertyMicrosoftCdoApptSequence ical.ComponentProperty = "X-MICROSOFT-CDO-APPT-SEQUENCE"
	// ComponentPropertyMicrosoftCdoAttendeeCriticalChange is the X-MICROSOFT-CDO-ATTENDEE-CRITICAL-CHANGE component property
	ComponentPropertyMicrosoftCdoAttendeeCriticalChange ical.ComponentProperty = "X-MICROSOFT-CDO-ATTENDEE-CRITICAL-CHANGE"
	// ComponentPropertyMicrosoftCdoBusystatus is the X-MICROSOFT-CDO-BUSYSTATUS component property
	ComponentPropertyMicrosoftCdoBusystatus ical.ComponentProperty = "X-MICROSOFT-CDO-BUSYSTATUS"
	// ComponentPropertyMicrosoftCdoImportance is the X-MICROSOFT-CDO-IMPORTANCE component property
	ComponentPropertyMicrosoftCdoImportance ical.ComponentProperty = "X-MICROSOFT-CDO-IMPORTANCE"
	// ComponentPropertyMicrosoftCdoInsttype is the X-MICROSOFT-CDO-INSTTYPE component property
	ComponentPropertyMicrosoftCdoInsttype ical.ComponentProperty = "X-MICROSOFT-CDO-INSTTYPE"
	// ComponentPropertyMicrosoftCdoIntendedstatus is the X-MICROSOFT-CDO-INTENDEDSTATUS component property
	ComponentPropertyMicrosoftCdoIntendedstatus ical.ComponentProperty = "X-MICROSOFT-CDO-INTENDEDSTATUS"
	// ComponentPropertyMicrosoftCdoOwnerapptid is the X-MICROSOFT-CDO-OWNERAPPTID component property
	ComponentPropertyMicrosoftCdoOwnerapptid ical.ComponentProperty = "X-MICROSOFT-CDO-OWNERAPPTID"
	// ComponentPropertyMicrosoftCdoOwnerCriticalChange is the X-MICROSOFT-CDO-OWNER-CRITICAL-CHANGE component property
	ComponentPropertyMicrosoftCdoOwnerCriticalChange ical.ComponentProperty = "X-MICROSOFT-CDO-OWNER-CRITICAL-CHANGE"
	// ComponentPropertyMicrosoftCdoReplytime is the X-MICROSOFT-CDO-REPLYTIME component property
	ComponentPropertyMicrosoftCdoReplytime ical.ComponentProperty = "X-MICROSOFT-CDO-REPLYTIME"
	// ComponentPropertyMicrosoftDisallowCounter is the X-MICROSOFT-DISALLOW-COUNTER component property
	ComponentPropertyMicrosoftDisallowCounter ical.ComponentProperty = "X-MICROSOFT-DISALLOW-COUNTER"
	// ComponentPropertyMicrosoftExdate is the X-MICROSOFT-EXDATE component property
	ComponentPropertyMicrosoftExdate ical.ComponentProperty = "X-MICROSOFT-EXDATE"
	// ComponentPropertyMicrosoftIsdraft is the X-MICROSOFT-ISDRAFT component property
	ComponentPropertyMicrosoftIsdraft ical.ComponentProperty = "X-MICROSOFT-ISDRAFT"
	// ComponentPropertyMicrosoftMsncalendarAlldayevent is the X-MICROSOFT-MSNCALENDAR-ALLDAYEVENT component property
	ComponentPropertyMicrosoftMsncalendarAlldayevent ical.ComponentProperty = "X-MICROSOFT-MSNCALENDAR-ALLDAYEVENT"
	// ComponentPropertyMicrosoftMsncalendarBusystatus is the X-MICROSOFT-MSNCALENDAR-BUSYSTATUS component property
	ComponentPropertyMicrosoftMsncalendarBusystatus ical.ComponentProperty = "X-MICROSOFT-MSNCALENDAR-BUSYSTATUS"
	// ComponentPropertyMicrosoftMsncalendarImportance is the X-MICROSOFT-MSNCALENDAR-IMPORTANCE component property
	ComponentPropertyMicrosoftMsncalendarImportance ical.ComponentProperty = "X-MICROSOFT-MSNCALENDAR-IMPORTANCE"
	// ComponentPropertyMicrosoftMsncalendarIntendedstatus is the X-MICROSOFT-MSNCALENDAR-INTENDEDSTATUS component property
	ComponentPropertyMicrosoftMsncalendarIntendedstatus ical.ComponentProperty = "X-MICROSOFT-MSNCALENDAR-INTENDEDSTATUS"
	// ComponentPropertyMicrosoftRrule is the X-MICROSOFT-RRULE component property
	ComponentPropertyMicrosoftRrule ical.ComponentProperty = "X-MICROSOFT-RRULE"
	// ComponentPropertyMsOlkAllowexterncheck is the X-MS-OLK-ALLOWEXTERNCHECK component property
	ComponentPropertyMsOlkAllowexterncheck ical.ComponentProperty = "X-MS-OLK-ALLOWEXTERNCHECK"
	// ComponentPropertyMsOlkApptlastsequence is the X-MS-OLK-APPTLASTSEQUENCE component property
	ComponentPropertyMsOlkApptlastsequence ical.ComponentProperty = "X-MS-OLK-APPTLASTSEQUENCE"
	// ComponentPropertyMsOlkApptseqtime is the X-MS-OLK-APPTSEQTIME component property
	ComponentPropertyMsOlkApptseqtime ical.ComponentProperty = "X-MS-OLK-APPTSEQTIME"
	// ComponentPropertyMsOlkAutofilllocation is the X-MS-OLK-AUTOFILLLOCATION component property
	ComponentPropertyMsOlkAutofilllocation ical.ComponentProperty = "X-MS-OLK-AUTOFILLLOCATION"
	// ComponentPropertyMsOlkAutostartcheck is the X-MS-OLK-AUTOSTARTCHECK component property
	ComponentPropertyMsOlkAutostartcheck ical.ComponentProperty = "X-MS-OLK-AUTOSTARTCHECK"
	// ComponentPropertyMsOlkCollaboratedoc is the X-MS-OLK-COLLABORATEDOC component property
	ComponentPropertyMsOlkCollaboratedoc ical.ComponentProperty = "X-MS-OLK-COLLABORATEDOC"
	// ComponentPropertyMsOlkConfcheck is the X-MS-OLK-CONFCHECK component property
	ComponentPropertyMsOlkConfcheck ical.ComponentProperty = "X-MS-OLK-CONFCHECK"
	// ComponentPropertyMsOlkConftype is the X-MS-OLK-CONFTYPE component property
	ComponentPropertyMsOlkConftype ical.ComponentProperty = "X-MS-OLK-CONFTYPE"
	// ComponentPropertyMsOlkDirectory is the X-MS-OLK-DIRECTORY component property
	ComponentPropertyMsOlkDirectory ical.ComponentProperty = "X-MS-OLK-DIRECTORY"
	// ComponentPropertyMsOlkMwsurl is the X-MS-OLK-MWSURL component property
	ComponentPropertyMsOlkMwsurl ical.ComponentProperty = "X-MS-OLK-MWSURL"
	// ComponentPropertyMsOlkNetshowurl is the X-MS-OLK-NETSHOWURL component property
	ComponentPropertyMsOlkNetshowurl ical.ComponentProperty = "X-MS-OLK-NETSHOWURL"
	// ComponentPropertyMsOlkOnlinepassword is the X-MS-OLK-ONLINEPASSWORD component property
	ComponentPropertyMsOlkOnlinepassword ical.ComponentProperty = "X-MS-OLK-ONLINEPASSWORD"
	// ComponentPropertyMsOlkOrgalias is the X-MS-OLK-ORGALIAS component property
	ComponentPropertyMsOlkOrgalias ical.ComponentProperty = "X-MS-OLK-ORGALIAS"
	// ComponentPropertyMsOlkOriginalend is the X-MS-OLK-ORIGINALEND component property
	ComponentPropertyMsOlkOriginalend ical.ComponentProperty = "X-MS-OLK-ORIGINALEND"
	// ComponentPropertyMsOlkOriginalstart is the X-MS-OLK-ORIGINALSTART component property
	ComponentPropertyMsOlkOriginalstart ical.ComponentProperty = "X-MS-OLK-ORIGINALSTART"
	// ComponentPropertyMsOlkSender is the X-MS-OLK-SENDER component property
	ComponentPropertyMsOlkSender ical.ComponentProperty = "X-MS-OLK-SENDER"
)

type propertySetter interface {
	SetProperty(property ical.ComponentProperty, value string, props ...ical.PropertyParameter)
}

// SetProperty allows extending the properties easily
func SetProperty(cal *ical.Calendar, property string, value string, params ...ical.PropertyParameter) {
	if cal == nil {
		return
	}
	// Fallback to directly modify CalendarProperties if no public SetProperty exists
	found := false
	for i := range cal.CalendarProperties {
		if cal.CalendarProperties[i].IANAToken == property {
			cal.CalendarProperties[i].Value = value
			cal.CalendarProperties[i].ICalParameters = map[string][]string{}
			for _, p := range params {
				k, v := p.KeyValue()
				cal.CalendarProperties[i].ICalParameters[k] = v
			}
			found = true
			break
		}
	}
	if !found {
		r := ical.CalendarProperty{
			BaseProperty: ical.BaseProperty{
				IANAToken:      property,
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
// SetCalend sets the X-CALEND property for the calendar
func SetCalend(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyCalend), value, params...)
}

// SetCalstart sets the X-CALSTART property for the calendar
func SetCalstart(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyCalstart), value, params...)
}

// SetClipend sets the X-CLIPEND property for the calendar
func SetClipend(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyClipend), value, params...)
}

// SetClipstart sets the X-CLIPSTART property for the calendar
func SetClipstart(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyClipstart), value, params...)
}

// SetMicrosoftCalscale sets the X-MICROSOFT-CALSCALE property for the calendar
func SetMicrosoftCalscale(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyMicrosoftCalscale), value, params...)
}

// SetMsOlkForceinspectoropen sets the X-MS-OLK-FORCEINSPECTOROPEN property for the calendar
func SetMsOlkForceinspectoropen(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyMsOlkForceinspectoropen), value, params...)
}

// SetMsWkhrdays sets the X-MS-WKHRDAYS property for the calendar
func SetMsWkhrdays(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyMsWkhrdays), value, params...)
}

// SetMsWkhrend sets the X-MS-WKHREND property for the calendar
func SetMsWkhrend(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyMsWkhrend), value, params...)
}

// SetMsWkhrstart sets the X-MS-WKHRSTART property for the calendar
func SetMsWkhrstart(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyMsWkhrstart), value, params...)
}

// SetOwner sets the X-OWNER property for the calendar
func SetOwner(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyOwner), value, params...)
}

// SetPrimaryCalendar sets the X-PRIMARY-CALENDAR property for the calendar
func SetPrimaryCalendar(cal *ical.Calendar, value string, params ...ical.PropertyParameter) {
	SetProperty(cal, string(PropertyPrimaryCalendar), value, params...)
}

// Component properties
// SetAltDesc sets the X-ALT-DESC property for a component
func SetAltDesc(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyAltDesc), value, params...)
}

// SetMicrosoftCdoAlldayevent sets the X-MICROSOFT-CDO-ALLDAYEVENT property for a component
func SetMicrosoftCdoAlldayevent(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoAlldayevent), value, params...)
}

// SetMicrosoftCdoApptSequence sets the X-MICROSOFT-CDO-APPT-SEQUENCE property for a component
func SetMicrosoftCdoApptSequence(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoApptSequence), value, params...)
}

// SetMicrosoftCdoAttendeeCriticalChange sets the X-MICROSOFT-CDO-ATTENDEE-CRITICAL-CHANGE property for a component
func SetMicrosoftCdoAttendeeCriticalChange(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoAttendeeCriticalChange), value, params...)
}

// SetMicrosoftCdoBusystatus sets the X-MICROSOFT-CDO-BUSYSTATUS property for a component
func SetMicrosoftCdoBusystatus(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoBusystatus), value, params...)
}

// SetMicrosoftCdoImportance sets the X-MICROSOFT-CDO-IMPORTANCE property for a component
func SetMicrosoftCdoImportance(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoImportance), value, params...)
}

// SetMicrosoftCdoInsttype sets the X-MICROSOFT-CDO-INSTTYPE property for a component
func SetMicrosoftCdoInsttype(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoInsttype), value, params...)
}

// SetMicrosoftCdoIntendedstatus sets the X-MICROSOFT-CDO-INTENDEDSTATUS property for a component
func SetMicrosoftCdoIntendedstatus(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoIntendedstatus), value, params...)
}

// SetMicrosoftCdoOwnerapptid sets the X-MICROSOFT-CDO-OWNERAPPTID property for a component
func SetMicrosoftCdoOwnerapptid(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoOwnerapptid), value, params...)
}

// SetMicrosoftCdoOwnerCriticalChange sets the X-MICROSOFT-CDO-OWNER-CRITICAL-CHANGE property for a component
func SetMicrosoftCdoOwnerCriticalChange(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoOwnerCriticalChange), value, params...)
}

// SetMicrosoftCdoReplytime sets the X-MICROSOFT-CDO-REPLYTIME property for a component
func SetMicrosoftCdoReplytime(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftCdoReplytime), value, params...)
}

// SetMicrosoftDisallowCounter sets the X-MICROSOFT-DISALLOW-COUNTER property for a component
func SetMicrosoftDisallowCounter(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftDisallowCounter), value, params...)
}

// SetMicrosoftExdate sets the X-MICROSOFT-EXDATE property for a component
func SetMicrosoftExdate(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftExdate), value, params...)
}

// SetMicrosoftIsdraft sets the X-MICROSOFT-ISDRAFT property for a component
func SetMicrosoftIsdraft(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftIsdraft), value, params...)
}

// SetMicrosoftMsncalendarAlldayevent sets the X-MICROSOFT-MSNCALENDAR-ALLDAYEVENT property for a component
func SetMicrosoftMsncalendarAlldayevent(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftMsncalendarAlldayevent), value, params...)
}

// SetMicrosoftMsncalendarBusystatus sets the X-MICROSOFT-MSNCALENDAR-BUSYSTATUS property for a component
func SetMicrosoftMsncalendarBusystatus(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftMsncalendarBusystatus), value, params...)
}

// SetMicrosoftMsncalendarImportance sets the X-MICROSOFT-MSNCALENDAR-IMPORTANCE property for a component
func SetMicrosoftMsncalendarImportance(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftMsncalendarImportance), value, params...)
}

// SetMicrosoftMsncalendarIntendedstatus sets the X-MICROSOFT-MSNCALENDAR-INTENDEDSTATUS property for a component
func SetMicrosoftMsncalendarIntendedstatus(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftMsncalendarIntendedstatus), value, params...)
}

// SetMicrosoftRrule sets the X-MICROSOFT-RRULE property for a component
func SetMicrosoftRrule(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMicrosoftRrule), value, params...)
}

// SetMsOlkAllowexterncheck sets the X-MS-OLK-ALLOWEXTERNCHECK property for a component
func SetMsOlkAllowexterncheck(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkAllowexterncheck), value, params...)
}

// SetMsOlkApptlastsequence sets the X-MS-OLK-APPTLASTSEQUENCE property for a component
func SetMsOlkApptlastsequence(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkApptlastsequence), value, params...)
}

// SetMsOlkApptseqtime sets the X-MS-OLK-APPTSEQTIME property for a component
func SetMsOlkApptseqtime(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkApptseqtime), value, params...)
}

// SetMsOlkAutofilllocation sets the X-MS-OLK-AUTOFILLLOCATION property for a component
func SetMsOlkAutofilllocation(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkAutofilllocation), value, params...)
}

// SetMsOlkAutostartcheck sets the X-MS-OLK-AUTOSTARTCHECK property for a component
func SetMsOlkAutostartcheck(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkAutostartcheck), value, params...)
}

// SetMsOlkCollaboratedoc sets the X-MS-OLK-COLLABORATEDOC property for a component
func SetMsOlkCollaboratedoc(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkCollaboratedoc), value, params...)
}

// SetMsOlkConfcheck sets the X-MS-OLK-CONFCHECK property for a component
func SetMsOlkConfcheck(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkConfcheck), value, params...)
}

// SetMsOlkConftype sets the X-MS-OLK-CONFTYPE property for a component
func SetMsOlkConftype(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkConftype), value, params...)
}

// SetMsOlkDirectory sets the X-MS-OLK-DIRECTORY property for a component
func SetMsOlkDirectory(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkDirectory), value, params...)
}

// SetMsOlkMwsurl sets the X-MS-OLK-MWSURL property for a component
func SetMsOlkMwsurl(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkMwsurl), value, params...)
}

// SetMsOlkNetshowurl sets the X-MS-OLK-NETSHOWURL property for a component
func SetMsOlkNetshowurl(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkNetshowurl), value, params...)
}

// SetMsOlkOnlinepassword sets the X-MS-OLK-ONLINEPASSWORD property for a component
func SetMsOlkOnlinepassword(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkOnlinepassword), value, params...)
}

// SetMsOlkOrgalias sets the X-MS-OLK-ORGALIAS property for a component
func SetMsOlkOrgalias(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkOrgalias), value, params...)
}

// SetMsOlkOriginalend sets the X-MS-OLK-ORIGINALEND property for a component
func SetMsOlkOriginalend(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkOriginalend), value, params...)
}

// SetMsOlkOriginalstart sets the X-MS-OLK-ORIGINALSTART property for a component
func SetMsOlkOriginalstart(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkOriginalstart), value, params...)
}

// SetMsOlkSender sets the X-MS-OLK-SENDER property for a component
func SetMsOlkSender(c ical.Component, value string, params ...ical.PropertyParameter) {
	SetComponentProperty(c, string(ComponentPropertyMsOlkSender), value, params...)
}
