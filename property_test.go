package ics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPropertyParse(t *testing.T) {
	tests := []struct {
		Name     string
		Input    string
		Expected func(output *BaseProperty) bool
	}{
		{Name: "ATTENDEE1", Input: "ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(output *BaseProperty) bool {
			return output.IANAToken == "ATTENDEE" && output.Value == "mailto:employee-A@example.com"
		}},
		{Name: "ATTENDEE2", Input: "ATTENDEE;RSVP=\"TRUE\";ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(output *BaseProperty) bool {
			return output.IANAToken == "ATTENDEE" && output.Value == "mailto:employee-A@example.com"
		}},
		{Name: "ATTENDEE - fail", Input: "ATTENDEE;RSVP=T\"RUE\";ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(output *BaseProperty) bool { return output == nil }},
		{Name: "X-ABC-MMSUBJ custom arg", Input: "X-ABC-MMSUBJ;VALUE=URI;FMTTYPE=audio/basic:http://www.example.org/mysubj.au", Expected: func(output *BaseProperty) bool {
			return output.IANAToken == "X-ABC-MMSUBJ" && output.Value == "audio/basic:http://www.example.org/mysubj.au"
		}},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			output := ParseProperty(ContentLine(test.Input))
			assert.True(t, test.Expected(output))
		})
	}
}
