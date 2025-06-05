# golang-ical
A  ICS / iCalendar parser and serialiser for Golang.

[![GoDoc](https://godoc.org/github.com/arran4/golang-ical?status.svg)](https://godoc.org/github.com/arran4/golang-ical)

Because the other libraries didn't quite do what I needed.

## How to parse an iCalendar file

```go
f, err := os.Open("calendar.ics")
if err != nil {
    log.Fatal(err)
}
cal, err := ics.ParseCalendar(f)
if err != nil {
    log.Fatal(err)
}
```

## How to parse from a URL

```go
cal, err := ics.ParseCalendarFromUrl("https://example.com/calendar.ics")
```

## How to read event details

```go
f, err := os.Open("calendar.ics")
if err != nil {
    log.Fatal(err)
}
cal, err := ics.ParseCalendar(f)
if err != nil {
    log.Fatal(err)
}
for _, evt := range cal.Events() {
    sum := evt.GetProperty(ics.ComponentPropertySummary)
    start, _ := evt.GetStartAt()
    if sum != nil {
        fmt.Printf("%s starts %s\n", sum.Value, start.Format(time.RFC3339))
    }
}
```

## How to create a simple event

```go
cal := ics.NewCalendar()
cal.SetProductId("-//My App//EN")
event := cal.AddEvent("1234@example.com")
event.SetDtStampTime(time.Now())
event.SetSummary("Dinner")
event.SetStartAt(time.Date(2024, 6, 1, 19, 0, 0, 0, time.UTC))
event.SetEndAt(time.Date(2024, 6, 1, 20, 0, 0, 0, time.UTC))
fmt.Print(cal.Serialize())
```

## How to add attendees and attachments

```go
event.SetOrganizer("organizer@example.com")
event.AddAttendee("guest@example.com",
    ics.CalendarUserTypeIndividual,
    ics.ParticipationStatusAccepted,
    ics.ParticipationRoleReqParticipant,
)
event.AddAttachmentURL("https://example.com/menu.pdf", "application/pdf")
```

## How to create a recurring event

```go
e := cal.AddEvent("weekly-meeting@example.com")
e.SetSummary("Weekly Meeting")
e.SetStartAt(time.Date(2024, 6, 3, 9, 0, 0, 0, time.UTC))
e.SetEndAt(time.Date(2024, 6, 3, 10, 0, 0, 0, time.UTC))
e.AddRrule("FREQ=WEEKLY;BYDAY=MO")
e.AddExdate("20240701T090000Z")
```
## How to create a to-do item

```go
cal := ics.NewCalendar()
cal.SetProductId("-//My App//EN")

todo := cal.AddTodo("finish-report@example.com")
todo.SetSummary("Finish quarterly report")
todo.SetDueAt(time.Date(2024, 6, 5, 17, 0, 0, 0, time.UTC))
todo.SetPercentComplete(50)
todo.SetPriority(1)
fmt.Print(cal.Serialize())
```

## How to add an alarm to an event

```go
e := cal.AddEvent("dentist@example.com")
e.SetSummary("Dentist Appointment")
e.SetStartAt(time.Date(2024, 6, 10, 15, 30, 0, 0, time.UTC))
e.SetEndAt(time.Date(2024, 6, 10, 16, 0, 0, 0, time.UTC))

alarm := e.AddAlarm()
alarm.SetAction(ics.ActionDisplay)
alarm.SetDescription("Time for your appointment")
alarm.SetTrigger("-PT15M")
```

## How to specify timezone information

```go
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
```


## Example programs

More complete programs can be found in the `examples` directory.
