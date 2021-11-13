package main

import (
	"Deliveroo/internal/parser"
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if l := len(args); l != 1 {
		log.Fatalf("1 argument expected, %d given", l)
	}

	ce, err := parser.NewCronExpression(args[0])
	if err != nil {
		log.Fatalf("error parsing expression: %s", err)
	}

	fmt.Print(ce.String())
}
