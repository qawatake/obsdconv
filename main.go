package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	Version = "x.y.z"
)

func main() {
	var flags flagBundle
	initFlags(flag.CommandLine, &flags)
	flag.Parse()
	if flags.ver {
		fmt.Printf("v%s\n", Version)
		return
	}
	if err := setFlags(flag.CommandLine, &flags); err != nil {
		log.Fatal(err)
	}
	if err := verifyFlags(&flags); err != nil {
		log.Fatal(err)
	}
	processor := newDefaultProcessor(&flags)
	if err := Walk(flags.src, flags.dst, processor); err != nil {
		log.Fatal(err)
	}
}
