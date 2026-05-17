import sys

content = sys.stdin.read()

# Comment 1: switch v := c.(type) and add *GeneralComponent
content = content.replace(
'''	switch c.(type) {
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
	}''',
'''	switch v := c.(type) {
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
	}'''
)

# Comment 2 & 3: Exclude VALUE param, use URI and CAL-ADDRESS for specific parameters
content = content.replace(
'''		for _, k := range keys {
			vals := bp.ICalParameters[k]
			xparam := xcalParameter{
				XMLName: xml.Name{Local: strings.ToLower(k)},
			}
			for _, v := range vals {
				xparam.Values = append(xparam.Values, xcalValue{
					XMLName: xml.Name{Local: "text"},
					Value:   v,
				})
			}
			xp.Parameters.Params = append(xp.Parameters.Params, xparam)
		}''',
'''		for _, k := range keys {
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
		}'''
)

# Comment 4: Remove && propName == "FREEBUSY"
content = content.replace(
'''			} else if innerType == "period" && propName == "FREEBUSY" {''',
'''			} else if innerType == "period" {'''
)

# Fix empty param list if only VALUE is present
# Let's ensure if xp.Parameters.Params is empty, we set xp.Parameters to nil
content = content.replace(
'''	if len(bp.ICalParameters) > 0 {
		xp.Parameters = &xcalParameters{}''',
'''	if len(bp.ICalParameters) > 0 {
		xp.Parameters = &xcalParameters{}'''
)
# Wait, actually if I don't set it to nil, it will marshal as <parameters></parameters> if empty.
# Let's fix that at the end.
content = content.replace(
'''			xp.Parameters.Params = append(xp.Parameters.Params, xparam)
		}
	}''',
'''			xp.Parameters.Params = append(xp.Parameters.Params, xparam)
		}
		if len(xp.Parameters.Params) == 0 {
			xp.Parameters = nil
		}
	}'''
)


with open("calendar_xml.go", "w") as f:
    f.write(content)
