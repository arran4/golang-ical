package microsoft

import (
	ical "github.com/arran4/golang-ical"
	"testing"
)

func TestCalendarProperties(t *testing.T) {
	cal := ical.NewCalendar()
	cal.SetXWRCalName("Test")
	SetMicrosoftCalscale(cal, "GREGORIAN")
	cal.SetXPublishedTTL("PT12H")

	foundCalname := false
	foundCalscale := false
	foundTtl := false

	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == string(ical.PropertyXWRCalName) {
			foundCalname = true
			if prop.Value != "Test" {
				t.Errorf("Expected Test, got %s", prop.Value)
			}
		}
		if prop.IANAToken == string(PropertyMicrosoftCalscale) {
			foundCalscale = true
			if prop.Value != "GREGORIAN" {
				t.Errorf("Expected GREGORIAN, got %s", prop.Value)
			}
		}
		if prop.IANAToken == string(ical.PropertyXPublishedTTL) {
			foundTtl = true
			if prop.Value != "PT12H" {
				t.Errorf("Expected PT12H, got %s", prop.Value)
			}
		}
	}

	if !foundCalname {
		t.Errorf("Property %s not found", ical.PropertyXWRCalName)
	}
	if !foundCalscale {
		t.Errorf("Property %s not found", PropertyMicrosoftCalscale)
	}
	if !foundTtl {
		t.Errorf("Property %s not found", ical.PropertyXPublishedTTL)
	}
}

func TestComponentProperties(t *testing.T) {
	event := ical.NewEvent("1")
	SetMicrosoftCdoAlldayevent(event, "TRUE")
	SetMicrosoftCdoBusystatus(event, "OOF")
	SetMicrosoftCdoImportance(event, "1")
	SetMicrosoftIsdraft(event, "TRUE")

	foundAllday := false
	foundBusy := false
	foundImp := false
	foundIsdraft := false

	for _, prop := range event.Properties {
		if prop.IANAToken == string(ComponentPropertyMicrosoftCdoAlldayevent) {
			foundAllday = true
			if prop.Value != "TRUE" {
				t.Errorf("Expected TRUE, got %s", prop.Value)
			}
		}
		if prop.IANAToken == string(ComponentPropertyMicrosoftCdoBusystatus) {
			foundBusy = true
			if prop.Value != "OOF" {
				t.Errorf("Expected OOF, got %s", prop.Value)
			}
		}
		if prop.IANAToken == string(ComponentPropertyMicrosoftCdoImportance) {
			foundImp = true
			if prop.Value != "1" {
				t.Errorf("Expected 1, got %s", prop.Value)
			}
		}
		if prop.IANAToken == string(ComponentPropertyMicrosoftIsdraft) {
			foundIsdraft = true
			if prop.Value != "TRUE" {
				t.Errorf("Expected TRUE, got %s", prop.Value)
			}
		}
	}

	if !foundAllday {
		t.Errorf("Property %s not found", ComponentPropertyMicrosoftCdoAlldayevent)
	}
	if !foundBusy {
		t.Errorf("Property %s not found", ComponentPropertyMicrosoftCdoBusystatus)
	}
	if !foundImp {
		t.Errorf("Property %s not found", ComponentPropertyMicrosoftCdoImportance)
	}
	if !foundIsdraft {
		t.Errorf("Property %s not found", ComponentPropertyMicrosoftIsdraft)
	}
}
