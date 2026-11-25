package main

import (
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
)

func main() {
	cal := ics.NewCalendar()
	cal.SetProductId("-//Example Corp//Basic Event//EN")
	event := cal.AddEvent("basic-event@example.com")
	event.SetSummary("Daily standup")
	event.SetStartAt(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC))
	event.SetEndAt(time.Date(2024, 6, 1, 9, 15, 0, 0, time.UTC))
	event.SetLocation("Meeting Room 1")
	fmt.Print(cal.Serialize())
}
