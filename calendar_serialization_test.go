//go:build go1.16
// +build go1.16

package ics

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/serialization/*.ics testdata/serialization/expected/*.ics
var serializationFixtures embed.FS

func TestCalendar_ReSerialization(t *testing.T) {
	testDir := "testdata/serialization"
	expectedDir := path.Join(testDir, "expected")
	testFiles, err := fs.Glob(serializationFixtures, testDir+"/*.ics")
	require.NoError(t, err)
	require.NotEmpty(t, testFiles)

	for _, fp := range testFiles {
		filename := path.Base(fp)
		t.Run(fmt.Sprintf("compare serialized -> deserialized -> serialized: %s", fp), func(t *testing.T) {
			//given
			originalSeriailizedCal, err := serializationFixtures.ReadFile(fp)
			require.NoError(t, err)

			//when
			deserializedCal, err := ParseCalendar(bytes.NewReader(originalSeriailizedCal))
			require.NoError(t, err)
			serializedCal := deserializedCal.Serialize(WithNewLineWindows)

			//then
			expectedCal, err := serializationFixtures.ReadFile(path.Join(expectedDir, filename))
			require.NoError(t, err)
			if diff := cmp.Diff(string(expectedCal), serializedCal); diff != "" {
				t.Error(diff)
			}
		})

		t.Run(fmt.Sprintf("compare deserialized -> serialized -> deserialized: %s", filename), func(t *testing.T) {
			//given
			loadIcsContent, err := serializationFixtures.ReadFile(path.Join(testDir, filename))
			require.NoError(t, err)
			originalDeserializedCal, err := ParseCalendar(bytes.NewReader(loadIcsContent))
			require.NoError(t, err)

			//when
			serializedCal := originalDeserializedCal.Serialize()
			deserializedCal, err := ParseCalendar(strings.NewReader(serializedCal))
			require.NoError(t, err)

			//then
			if diff := cmp.Diff(originalDeserializedCal, deserializedCal, cmpopts.IgnoreUnexported(Calendar{}, ComponentBase{})); diff != "" {
				t.Error(diff)
			}
		})
	}
}
