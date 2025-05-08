package main

import (
	"log"

	"github.com/dmytrii/termai/pkg/cmd/parser"
)

func main() {
	newParser := parser.New()
	newParser.Init()

	if err := newParser.Execute(); err != nil {
		log.Fatal(err)
	}
} 