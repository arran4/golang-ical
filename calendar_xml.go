package ics

import (
	"encoding/xml"
	"sort"
	"strings"
)

// XML formatting types according to RFC 6321

type xcalIcalendar struct {
	XMLName   xml.Name      `xml:"urn:ietf:params:xml:ns:icalendar-2.0 icalendar"`
	Vcalendar xcalVcalendar `xml:"vcalendar"`
}

type xcalVcalendar struct {
	Properties xcalProperties `xml:"properties"`
	Components xcalComponents `xml:"components"`
}

type xcalProperties struct {
	Props []xcalProperty `xml:",any"`
}

type xcalComponents struct {
	Comps []xcalComponent `xml:",any"`
}

type xcalComponent struct {
	XMLName    xml.Name
	Properties xcalProperties  `xml:"properties"`
	Components *xcalComponents `xml:"components,omitempty"`
}

type xcalProperty struct {
	XMLName    xml.Name
	Parameters *xcalParameters `xml:"parameters,omitempty"`
	// Inner elements depend on value type and special cases
	Inner []xcalValue `xml:",any"`
}

type xcalParameters struct {
	Params []xcalParameter `xml:",any"`
}

type xcalParameter struct {
	XMLName xml.Name
	Values  []xcalValue `xml:",any"`
}

type xcalValue struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// MarshalXML implements xml.Marshaler for Calendar.
func (cal *Calendar) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	vcal := xcalVcalendar{
		Properties: xcalProperties{},
		Components: xcalComponents{},
	}

	for _, p := range cal.CalendarProperties {
		vcal.Properties.Props = append(vcal.Properties.Props, p.BaseProperty.toXcalProperty())
	}

	for _, c := range cal.Components {
		vcal.Components.Comps = append(vcal.Components.Comps, componentToXcal(c))
	}

	wrapper := xcalIcalendar{
		Vcalendar: vcal,
	}

	// We override the start element to use our specific element name
	start.Name = xml.Name{Space: "urn:ietf:params:xml:ns:icalendar-2.0", Local: "icalendar"}

	return e.EncodeElement(wrapper, start)
}

func componentToXcal(c Component) xcalComponent {
	cb := c.UnknownPropertiesIANAProperties()

	compType := ""
	switch v := c.(type) {
	case *VEvent:
		compType = "vevent"
	case *VTodo:
		compType = "vtodo"
	case *VJournal:
		compType = "vjournal"
	case *VBusy:
		compType = "vfreebusy"
	case *VTimezone:
		compType = "vtimezone"
	case *VAlarm:
		compType = "valarm"
	case *Standard:
		compType = "standard"
	case *Daylight:
		compType = "daylight"
	case *GeneralComponent:
		compType = strings.ToLower(v.Token)
	default:
		compType = "unknown"
	}

	xc := xcalComponent{
		XMLName: xml.Name{Local: compType},
	}

	for _, p := range cb {
		xc.Properties.Props = append(xc.Properties.Props, p.BaseProperty.toXcalProperty())
	}

	subcomps := c.SubComponents()
	if len(subcomps) > 0 {
		xc.Components = &xcalComponents{}
		for _, sc := range subcomps {
			xc.Components.Comps = append(xc.Components.Comps, componentToXcal(sc))
		}
	}

	return xc
}

func (bp BaseProperty) toXcalProperty() xcalProperty {
	xp := xcalProperty{
		XMLName: xml.Name{Local: strings.ToLower(bp.IANAToken)},
	}

	if len(bp.ICalParameters) > 0 {
		xp.Parameters = &xcalParameters{}
		// Sort keys to ensure deterministic output
		keys := make([]string, 0, len(bp.ICalParameters))
		for k := range bp.ICalParameters {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if strings.ToUpper(k) == "VALUE" {
				continue
			}
			vals := bp.ICalParameters[k]
			xparam := xcalParameter{
				XMLName: xml.Name{Local: strings.ToLower(k)},
			}

			paramValType := "text"
			switch strings.ToUpper(k) {
			case "ALTREP", "DIR":
				paramValType = "uri"
			case "MEMBER", "DELEGATED-FROM", "DELEGATED-TO", "SENT-BY":
				paramValType = "cal-address"
			}
			for _, v := range vals {
				xparam.Values = append(xparam.Values, xcalValue{
					XMLName: xml.Name{Local: paramValType},
					Value:   v,
				})
			}
			xp.Parameters.Params = append(xp.Parameters.Params, xparam)
		}
		if len(xp.Parameters.Params) == 0 {
			xp.Parameters = nil
		}
	}

	valType := strings.ToLower(string(bp.GetValueType()))
	propName := strings.ToUpper(bp.IANAToken)

	if valType == "" || valType == "unknown" {
		valType = "text"
	}

	if propName == "GEO" {
		parts := strings.SplitN(bp.Value, ";", 2)
		lat := parts[0]
		lon := ""
		if len(parts) > 1 {
			lon = parts[1]
		}
		xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: "latitude"}, Value: lat})
		xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: "longitude"}, Value: lon})
	} else if propName == "REQUEST-STATUS" {
		parts := strings.Split(bp.Value, ";")
		if len(parts) > 0 {
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: "code"}, Value: parts[0]})
		}
		if len(parts) > 1 {
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: "description"}, Value: parts[1]})
		}
		if len(parts) > 2 {
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: "data"}, Value: parts[2]})
		}
	} else if propName == "CATEGORIES" || propName == "RESOURCES" || propName == "FREEBUSY" || propName == "EXDATE" || propName == "RDATE" {
		parts := strings.Split(bp.Value, ",")
		for _, part := range parts {
			// For EXDATE and RDATE, the value type might be DATE or DATE-TIME.
			// The original ICS might specify this via VALUE=DATE parameter. Let's rely on valType.
			innerType := valType
			innerVal := part
			if innerType == "date-time" {
				innerVal = formatXcalDateTime(part)
			} else if innerType == "date" {
				innerVal = formatXcalDate(part)
			} else if innerType == "period" {
				// Period formatting:
				periodParts := strings.SplitN(part, "/", 2)
				if len(periodParts) == 2 {
					p1 := periodParts[0]
					p2 := periodParts[1]
					p1 = formatXcalDateTime(p1)
					if strings.Contains(p2, "P") {
						// It's a duration, no formatting needed for duration usually, or maybe it does?
					} else {
						p2 = formatXcalDateTime(p2)
					}
					innerVal = p1 + "/" + p2
				}
			}
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: innerType}, Value: innerVal})
		}
	} else {
		if valType == "" {
			valType = "unknown"
		}

		switch valType {
		case "date-time":
			formatted := formatXcalDateTime(bp.Value)
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: valType}, Value: formatted})
		case "date":
			formatted := formatXcalDate(bp.Value)
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: valType}, Value: formatted})
		case "cal-address":
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: "cal-address"}, Value: bp.Value})
		case "utc-offset":
			formatted := formatXcalUtcOffset(bp.Value)
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: valType}, Value: formatted})
		default:
			xp.Inner = append(xp.Inner, xcalValue{XMLName: xml.Name{Local: valType}, Value: bp.Value})
		}
	}

	return xp
}

func formatXcalDateTime(val string) string {
	if len(val) == 15 {
		return val[0:4] + "-" + val[4:6] + "-" + val[6:11] + ":" + val[11:13] + ":" + val[13:]
	} else if len(val) == 16 {
		return val[0:4] + "-" + val[4:6] + "-" + val[6:11] + ":" + val[11:13] + ":" + val[13:]
	}
	return val
}

func formatXcalDate(val string) string {
	if len(val) == 8 {
		return val[0:4] + "-" + val[4:6] + "-" + val[6:8]
	}
	return val
}

func formatXcalUtcOffset(val string) string {
	if len(val) == 5 {
		return val[0:3] + ":" + val[3:5]
	} else if len(val) == 7 {
		return val[0:3] + ":" + val[3:5] + ":" + val[5:7]
	}
	return val
}
