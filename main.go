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
	// flags を設定
	flags := new(flagBundle)
	initFlags(flag.CommandLine, flags)
	flag.Parse()
	setFlags(flag.CommandLine, flags)

	// main 部分
	err := run(Version, flags, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func run(version string, flags *flagBundle, w io.Writer) (err error) {
	if flags.ver {
		fmt.Fprintf(w, "v%s\n", version)
		return nil
	}
	if err := verifyFlags(flags); err != nil {
		return err
	}
	processor := newDefaultProcessor(flags)
	skipper, err := process.NewSkipper(filepath.Join(flags.src, DEFAULT_IGNORE_FILE_NAME))
	if err != nil {
		return err
	}
	if err := process.Walk(flags.src, flags.dst, skipper, processor); err != nil {
		return err
	}
	return nil
}
