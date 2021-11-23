package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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
	// config を設定
	config := new(configuration)
	initFlags(flag.CommandLine, config)
	flag.Parse()
	setConfig(flag.CommandLine, config)

	// main 部分
	err := run(Version, config, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func run(version string, config *configuration, w io.Writer) (err error) {
	if config.ver {
		fmt.Fprintf(w, "v%s\n", version)
		return nil
	}
	if err := verifyConfig(config); err != nil {
		return err
	}
	skipper, err := process.NewSkipper(filepath.Join(config.src, DEFAULT_IGNORE_FILE_NAME))
	if err != nil {
		return err
	}
	processor, err := newDefaultProcessor(config)
	if err != nil {
		return err
	}
	if err := process.Walk(config.src, config.dst, skipper, processor); err != nil {
		return err
	}
	return nil
}
