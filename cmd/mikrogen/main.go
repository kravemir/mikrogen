package main

import (
	"flag"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/kravemir/mikrogen"
	"github.com/mitchellh/mapstructure"
	"os"
)

var printHelp bool

func init() {
	flag.BoolVar(&printHelp, "h", false, "Print usage (shorthand)")
	flag.BoolVar(&printHelp, "help", false, "Print usage")
}

func main() {
	flag.Parse()

	if printHelp {
		flag.Usage()
		return
	}

	filename := loadArguments()

	k := koanf.New(".")
	err := k.Load(file.Provider(filename), toml.Parser())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read config: %s\n", err.Error())
		os.Exit(1)
	}

	var c mikrogen.Configuration
	err = k.UnmarshalWithConf("", nil, koanf.UnmarshalConf{
		Tag:       "",
		FlatPaths: false,
		DecoderConfig: &mapstructure.DecoderConfig{
			ErrorUnused: true,
			Result:      &c,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to map config: %s\n", err.Error())
		os.Exit(1)
	}

	result := mikrogen.Generate(c)
	fmt.Print(result)
}

func loadArguments() string {
	args := flag.Args()
	switch len(args) {
	case 1:
		return args[0]
	default:
		fmt.Printf("Incorrect number of arguments passed, required 1 argument: <config_file>")
		os.Exit(2)
	}
	return ""
}
