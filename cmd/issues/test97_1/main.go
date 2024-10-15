package main

import (
	"fmt"
	ics "github.com/arran4/golang-ical"
	"net/url"
)

func main() {
	i := ics.NewCalendarFor("Mozilla.org/NONSGML Mozilla Calendar V1.1")
	tz := i.AddTimezone("Europe/Berlin")
	tz.AddProperty(ics.ComponentPropertyExtended("TZINFO"), "Europe/Berlin[2024a]")
	tzstd := tz.AddStandard()
	tzstd.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetto), "+010000")
	tzstd.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetfrom), "+005328")
	tzstd.AddProperty(ics.ComponentProperty(ics.PropertyTzname), "Europe/Berlin(STD)")
	tzstd.AddProperty(ics.ComponentProperty(ics.PropertyDtstart), "18930401T000000")
	tzstd.AddProperty(ics.ComponentProperty(ics.PropertyRdate), "18930401T000000")
	vEvent := i.AddEvent("d23cef0d-9e58-43c4-9391-5ad8483ca346")
	vEvent.AddProperty(ics.ComponentPropertyCreated, "20240929T120640Z")
	vEvent.AddProperty(ics.ComponentPropertyLastModified, "20240929T120731Z")
	vEvent.AddProperty(ics.ComponentPropertyDtstamp, "20240929T120731Z")
	vEvent.AddProperty(ics.ComponentPropertySummary, "Test Event")
	vEvent.AddProperty(ics.ComponentPropertyDtStart, "20240929T144500", ics.WithTZID("Europe/Berlin"))
	vEvent.AddProperty(ics.ComponentPropertyDtEnd, "20240929T154500", ics.WithTZID("Europe/Berlin"))
	vEvent.AddProperty(ics.ComponentPropertyTransp, "OPAQUE")
	vEvent.AddProperty(ics.ComponentPropertyLocation, "Github")
	uri := &url.URL{
		Scheme: "data",
		Opaque: "text/html,I%20want%20a%20custom%20linkout%20for%20Thunderbird.%3Cbr%3EThis%20is%20the%20Github%20%3Ca%20href%3D%22https%3A%2F%2Fgithub.com%2Farran4%2Fgolang-ical%2Fissues%2F97%22%3EIssue%3C%2Fa%3E.",
	}
	vEvent.AddProperty(ics.ComponentPropertyDescription, `"I want a custom linkout for Thunderbird.\nThis is the Github Issue.`, ics.WithAlternativeRepresentation(uri))
	fmt.Println(i.Serialize())
}
