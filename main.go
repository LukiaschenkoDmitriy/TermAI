package main

import (
	"log"

	"github.com/LukiaschenkoDmitriy/TermAI/pkg/cmd/parser"
)

func main() {
	newParser := parser.New()
	newParser.Init()

	if err := newParser.Execute(); err != nil {
		log.Fatal(err)
	}
} 