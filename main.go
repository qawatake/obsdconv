package main

import (
	"flag"
	"fmt"
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
	versionText, bufferredErrs, err := run(Version, config)
	if err != nil {
		log.Fatal(err)
	}
	if versionText != "" {
		fmt.Println(versionText)
		return
	}
	for _, err := range bufferredErrs {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run(version string, config *configuration) (verionText string, bufferredErrs []error, err error) {
	if config.ver {
		return fmt.Sprintf("v%s", version), nil, nil
	}
	if err := verifyConfig(config); err != nil {
		return "", nil, err
	}
	skipper, err := process.NewSkipper(filepath.Join(config.src, DEFAULT_IGNORE_FILE_NAME))
	if err != nil {
		return "", nil, err
	}
	processor, err := newDefaultProcessor(config)
	if err != nil {
		return "", nil, err
	}
	if err := process.Walk(config.src, config.dst, skipper, processor); err != nil {
		return "", nil, err
	}
	return "", processor.errbuf, nil
}
