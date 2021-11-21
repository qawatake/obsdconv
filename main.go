package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/qawatake/obsdconv/process"
)

var (
	Version = "x.y.z"
)

const (
	DEFAULT_IGNORE_FILE_NAME = ".obsdconvignore"
)

func main() {
	version, err := run(flag.CommandLine)
	if err != nil {
		log.Fatal(err)
	}
	if version != "" {
		fmt.Println(version)
	}
}

func run(flagset *flag.FlagSet) (version string, err error) {
	flags := new(flagBundle)
	initFlags(flag.CommandLine, flags)
	flag.Parse()
	if flags.ver {
		return fmt.Sprintf("v%s\n", Version), nil
	}
	setFlags(flag.CommandLine, flags)
	if err := verifyFlags(flags); err != nil {
		return "", err
	}
	processor := newDefaultProcessor(flags)
	skipper, err := process.NewSkipper(filepath.Join(flags.src, DEFAULT_IGNORE_FILE_NAME))
	if err != nil {
		return "", err
	}
	if err := process.Walk(flags.src, flags.dst, skipper, processor); err != nil {
		return "", err
	}
	return "", nil
}
