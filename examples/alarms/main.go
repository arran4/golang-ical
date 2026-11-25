package main

import (
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
)

func main() {
	cal := ics.NewCalendar()
	cal.SetProductId("-//Example Corp//Alarm Example//EN")

	e := cal.AddEvent("dentist@example.com")
	e.SetSummary("Dentist Appointment")
	e.SetStartAt(time.Date(2024, 6, 10, 15, 30, 0, 0, time.UTC))
	e.SetEndAt(time.Date(2024, 6, 10, 16, 0, 0, 0, time.UTC))

	alarm := e.AddAlarm()
	alarm.SetAction(ics.ActionDisplay)
	alarm.SetDescription("Time for your appointment")
	alarm.SetTrigger("-PT15M")

	fmt.Print(cal.Serialize())
}
