package main

import (
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
)

func main() {
	cal := ics.NewCalendar()
	cal.SetProductId("-//Example Corp//TODO//EN")

	todo := cal.AddTodo("finish-report@example.com")
	todo.SetSummary("Finish quarterly report")
	todo.SetDueAt(time.Date(2024, 6, 5, 17, 0, 0, 0, time.UTC))
	todo.SetPercentComplete(50)
	todo.SetPriority(1)

	alarm := todo.AddAlarm()
	alarm.SetAction(ics.ActionDisplay)
	alarm.SetDescription("Task due soon")
	alarm.SetTrigger("-PT30M")

	fmt.Print(cal.Serialize())
}
