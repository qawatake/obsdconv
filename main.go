package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/qawatake/obsdconv/convert"
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
	convert.PrepareVault(flags.src)
	if err := cwalk(&flags); err != nil {
		log.Fatal(err)
	}
}
