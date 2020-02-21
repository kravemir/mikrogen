package mikrogen

import (
	"fmt"
	"gotest.tools/assert"
	"io/ioutil"
	"path"
	"runtime"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name          string
		configuration Configuration
		want          string
	}{
		{
			"01",
			Configuration{
				IdentifierPrefix: "BlockWeb",

				DNSBlockedAddresses: []string{
					"facebook.com",
					"static.facebook.com",
					"www.facebook.com",
					"api.facebook.com",
					"edge-mqtt.facebook.com",
					"messenger.com",
					"www.messenger.com",
				},

				TLSBlockedAddresses: []string{
					"facebook.com",
					"*.facebook.com",
					"messenger.com",
					"*.messenger.com",
					"fbcdn.net",
					"*.fbcdn.net",
				},

				DisableIntervals: []Interval{
					{
						"08:30:00",
						"09:08:00",
					},
					{
						"14:00:00",
						"15:05:00",
					},
					{
						"18:00:00",
						"19:03:00",
					},
				},
			},
			loadFile("output_01.txt"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.configuration)

			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func loadFile(name string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := path.Dir(currentFile)
	filepath := path.Join(dir, name)
	result, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to load %s: %w", filepath, err))
	}
	return string(result)
}
