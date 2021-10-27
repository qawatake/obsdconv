package main

import (
	"flag"
	"log"
)

func main() {
	var flags flagBundle
	initFlags(flag.CommandLine, &flags)
	flag.Parse()
	if err := setFlags(flag.CommandLine, &flags); err != nil {
		log.Fatal(err)
	}
	if err := walk(&flags); err != nil {
		log.Fatal(err)
	}
}
