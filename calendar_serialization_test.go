//go:build go1.18
// +build go1.18

package ics

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestCalendar_ReSerialization(t *testing.T) {
	testDir := "testdata/serialization"
	expectedDir := filepath.Join(testDir, "expected")
	actualDir := filepath.Join(testDir, "actual")

	testFileNames := []string{
		"input1.ics",
		"input2.ics",
		"input3.ics",
		"input4.ics",
		"input5.ics",
		"input6.ics",
		"input7.ics",
	}

	for _, filename := range testFileNames {
		t.Run(fmt.Sprintf("compare serialized -> deserialized -> serialized: %s", filename), func(t *testing.T) {
			//given
			originalSeriailizedCal, err := os.ReadFile(filepath.Join(testDir, filename))
			require.NoError(t, err)

			//when
			deserializedCal, err := ParseCalendar(bytes.NewReader(originalSeriailizedCal))
			serializedCal := deserializedCal.Serialize()

			//then
			expectedCal, _ := os.ReadFile(filepath.Join(expectedDir, filename))
			if diff := cmp.Diff(string(expectedCal), serializedCal); diff != "" {
				os.MkdirAll(actualDir, 0755)
				os.WriteFile(filepath.Join(actualDir, filename), []byte(serializedCal), 0644)
				t.Error(diff)
			}
		})

		t.Run(fmt.Sprintf("compare deserialized -> serialized -> deserialized: %s", filename), func(t *testing.T) {
			//given
			loadIcsContent, err := os.ReadFile(filepath.Join(testDir, filename))
			require.NoError(t, err)
			originalDeserializedCal, err := ParseCalendar(bytes.NewReader(loadIcsContent))

			//when
			serializedCal := originalDeserializedCal.Serialize()
			deserializedCal, err := ParseCalendar(strings.NewReader(serializedCal))
			require.NoError(t, err)

			//then
			if diff := cmp.Diff(originalDeserializedCal, deserializedCal); diff != "" {
				t.Error(diff)
			}
		})
	}
}
