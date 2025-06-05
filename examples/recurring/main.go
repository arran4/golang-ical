package main

import (
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
)

func main() {
	cal := ics.NewCalendar()
	cal.SetProductId("-//Example Corp//Recurring Event//EN")

	e := cal.AddEvent("weekly-meeting@example.com")
	e.SetSummary("Weekly Meeting")
	e.SetDescription("Discuss project progress")
	e.SetStartAt(time.Date(2024, 6, 3, 9, 0, 0, 0, time.UTC))
	e.SetEndAt(time.Date(2024, 6, 3, 10, 0, 0, 0, time.UTC))

	e.AddRrule("FREQ=WEEKLY;BYDAY=MO")
	e.AddExdate("20240701T090000Z")

	e.SetOrganizer("manager@example.com")
	e.AddAttendee("dev@example.com",
		ics.CalendarUserTypeIndividual,
		ics.ParticipationStatusNeedsAction,
		ics.ParticipationRoleReqParticipant,
	)
	e.AddAttachmentURL("https://example.com/agenda.pdf", "application/pdf")

	fmt.Print(cal.Serialize())
}
