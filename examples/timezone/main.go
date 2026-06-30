package main

import (
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
)

func main() {
	cal := ics.NewCalendar()
	cal.SetProductId("-//Example Corp//Timezone Example//EN")

	tz := cal.AddTimezone("America/New_York")
	std := tz.AddStandard()
	std.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetto), "-0500")
	std.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetfrom), "-0400")
	std.AddProperty(ics.ComponentProperty(ics.PropertyDtstart), "19701101T020000")

	dst := &ics.Daylight{}
	dst.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetto), "-0400")
	dst.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetfrom), "-0500")
	dst.AddProperty(ics.ComponentProperty(ics.PropertyDtstart), "19700308T020000")
	tz.Components = append(tz.Components, dst)

	e := cal.AddEvent("meeting@example.com")
	e.SetSummary("Morning Meeting")
	e.SetDtStampTime(time.Now())
	e.SetStartAt(time.Date(2024, 6, 11, 9, 0, 0, 0, time.FixedZone("EDT", -4*3600)))
	e.SetEndAt(time.Date(2024, 6, 11, 10, 0, 0, 0, time.FixedZone("EDT", -4*3600)))
	e.SetLocation("New York Office")
	e.SetProperty(ics.ComponentPropertyTzid, "America/New_York")

	fmt.Print(cal.Serialize())
}
