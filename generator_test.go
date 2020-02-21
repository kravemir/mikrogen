package mikrogen

import (
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
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
			loadConfig("sample_01.in.toml"),
			loadFile("sample_01.out"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.configuration)

			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func loadConfig(filename string) Configuration {
	k := koanf.New(".")
	err := k.Load(file.Provider(getLocalFile(filename)), toml.Parser())
	if err != nil {
		panic(err)
	}

	var c Configuration
	err = k.Unmarshal("", &c)
	if err != nil {
		panic(err)
	}
	return c
}

func loadFile(name string) string {
	filepath := getLocalFile(name)
	result, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to load %s: %w", filepath, err))
	}
	return string(result)
}

func getLocalFile(name string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := path.Dir(currentFile)
	filepath := path.Join(dir, name)
	return filepath
}
