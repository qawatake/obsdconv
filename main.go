package main

import (
	"flag"
	"fmt"
	"log"
)

const (
	FLAG_SOURCE         = "src"
	FLAG_DESTINATION    = "dst"
	FLAG_REMOVE_TAGS    = "rmtag"
	FLAG_COPY_TAGS      = "cptag"
	FLAG_COPY_TITLE     = "title"
	FLAG_COPY_ALIASES   = "alias"
	FLAG_CONVERT_LINKS  = "link"
	FLAG_REMOVE_COMMENT = "cmmt"
	FLAG_OBSIDIAN_USAGE = "obs"
	FLAG_COMMON_USAGE   = "cmmn"
)

type flagBundle struct {
	src   string
	dst   string
	rmtag bool
	cptag bool
	title bool
	alias bool
	link  bool
	cmmt  bool
	obs   bool
	cmmn  bool
}

var (
	ErrFlagSourceNotSet      = fmt.Errorf("flag %s was not set", FLAG_SOURCE)
	ErrFlagDestinationNotSet = fmt.Errorf("flag %s was not set", FLAG_DESTINATION)
)

var flags flagBundle

func init() {
	initFlags(flag.CommandLine, &flags)
}

func main() {
	flag.Parse()
	if err := setFlags(flag.CommandLine, &flags); err != nil {
		log.Fatal(err)
	}
	if err := walk(&flags); err != nil {
		log.Fatal(err)
	}
}
