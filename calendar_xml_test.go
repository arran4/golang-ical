package ics

import (
	"encoding/xml"
	"testing"
)

func TestCalendarXMLMarshal(t *testing.T) {
	cal := NewCalendar()
	cal.SetMethod(MethodRequest)
	cal.SetProductId("-//Example Inc.//Example Calendar//EN")

	event := cal.AddEvent("4088E990AD89CB3DBB484909")
	event.SetSummary("Planning meeting")

	dtstamp, _ := ParseProperty(ContentLine("DTSTAMP:20080205T191224Z"))
	event.Properties = append(event.Properties, IANAProperty{*dtstamp})

	dtstart, _ := ParseProperty(ContentLine("DTSTART;VALUE=DATE:20081006"))
	event.Properties = append(event.Properties, IANAProperty{*dtstart})

	b, err := xml.MarshalIndent(cal, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	expectedXML := `<icalendar xmlns="urn:ietf:params:xml:ns:icalendar-2.0">
  <vcalendar>
    <properties>
      <version>
        <text>2.0</text>
      </version>
      <prodid>
        <text>-//Example Inc.//Example Calendar//EN</text>
      </prodid>
      <method>
        <text>REQUEST</text>
      </method>
    </properties>
    <components>
      <vevent>
        <properties>
          <uid>
            <text>4088E990AD89CB3DBB484909</text>
          </uid>
          <summary>
            <text>Planning meeting</text>
          </summary>
          <dtstamp>
            <date-time>2008-02-05T19:12:24Z</date-time>
          </dtstamp>
          <dtstart>
            <date>2008-10-06</date>
          </dtstart>
        </properties>
      </vevent>
    </components>
  </vcalendar>
</icalendar>`

	if string(b) != expectedXML {
		t.Errorf("XML marshal mismatch.\nExpected:\n%s\nGot:\n%s", expectedXML, string(b))
	}
}

func TestCalendarXMLMarshalComponent(t *testing.T) {
	cal := NewCalendar()
	cal.SetMethod(MethodRequest)
	cal.SetProductId("-//Example Inc.//Example Calendar//EN")

	event := cal.AddEvent("4088E990AD89CB3DBB484909")
	event.SetSummary("Planning meeting")

	geo, _ := ParseProperty(ContentLine("GEO:37.386013;-122.082932"))
	event.Properties = append(event.Properties, IANAProperty{*geo})

	reqStatus, _ := ParseProperty(ContentLine("REQUEST-STATUS:2.0;Success"))
	event.Properties = append(event.Properties, IANAProperty{*reqStatus})

	categories, _ := ParseProperty(ContentLine("CATEGORIES:APPOINTMENT,EDUCATION"))
	event.Properties = append(event.Properties, IANAProperty{*categories})

	b, err := xml.MarshalIndent(cal, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	expectedXML := `<icalendar xmlns="urn:ietf:params:xml:ns:icalendar-2.0">
  <vcalendar>
    <properties>
      <version>
        <text>2.0</text>
      </version>
      <prodid>
        <text>-//Example Inc.//Example Calendar//EN</text>
      </prodid>
      <method>
        <text>REQUEST</text>
      </method>
    </properties>
    <components>
      <vevent>
        <properties>
          <uid>
            <text>4088E990AD89CB3DBB484909</text>
          </uid>
          <summary>
            <text>Planning meeting</text>
          </summary>
          <geo>
            <latitude>37.386013</latitude>
            <longitude>-122.082932</longitude>
          </geo>
          <request-status>
            <code>2.0</code>
            <description>Success</description>
          </request-status>
          <categories>
            <text>APPOINTMENT</text>
            <text>EDUCATION</text>
          </categories>
        </properties>
      </vevent>
    </components>
  </vcalendar>
</icalendar>`

	if string(b) != expectedXML {
		t.Errorf("XML marshal mismatch.\nExpected:\n%s\nGot:\n%s", expectedXML, string(b))
	}
}
