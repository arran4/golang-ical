# golang-ical
A  ICS / ICal parser and serialiser for Golang.

Because the other libraries didn't quite do what I needed.

Usage, parsing:
```
    cal, err := ParseCalendar(strings.NewReader(input))

```

Creating:
```
  cal := ics.NewCalendar()
  cal.SetMethod(ics.MethodRequest)
  cal.SetCalscale("GREGORIAN")
  cal.SetName("Name")
  cal.SetXWRCalName("Name")
  cal.SetDescription("Description")
  cal.SetXWRCalDesc("Description")
  cal.SetXWRTimezone("UTC")
  
  event := cal.AddEvent(fmt.Sprintf("id@domain", p.SessionKey.IntID()))
  event.SetCreatedTime(time.Now())
  event.SetDtStampTime(time.Now())
  event.SetModifiedAt(time.Now())
  event.SetStartAt(time.Now())
  event.SetEndAt(time.Now())
  event.SetSummary("Summary")
  event.SetLocation("Address")
  event.SetDescription("Description")
  event.SetURL("https://URL/")
  event.SetOrganizer("sender@domain", ics.WithCN("This Machine"))
  event.AddAttendee("reciever or participant", ics.CalendarUserTypeIndividual, ics.ParticipationStatusNeedsAction, ics.ParticipationRoleReqParticipant, ics.WithRSVP(true))
  event.AddRrule("FREQ=WEEKLY;COUNT=10;WKST=SU;BYDAY=MO,WE,FR")
  
  alarm := event.AddAlarm()
  alarm.SetAction(ics.ActionDisplay)
  alarm.SetTrigger("-PT10M")
  
  return cal.Serialize()
```

Helper methods created as needed feel free to send a P.R. with more.
