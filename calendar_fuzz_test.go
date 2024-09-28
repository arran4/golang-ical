//go:build go1.18
// +build go1.18

package ics

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzParseCalendar(f *testing.F) {
	ics, err := os.ReadFile("testdata/timeparsing.ics")
	require.NoError(f, err)
	f.Add(ics)
	f.Fuzz(func(t *testing.T, ics []byte) {
		_, err := ParseCalendar(bytes.NewReader(ics))
		t.Log(err)
	})
}
