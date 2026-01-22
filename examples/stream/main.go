package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/cacack/gedcom-go/parser"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: stream <path-to-gedcom>")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer f.Close()

	p := parser.NewParser()
	scanner := bufio.NewScanner(f)
	scanner.Split(parser.ScanGEDCOMLines)

	individualCount := 0
	handleLine := func(line *parser.Line) error {
		if line.Level == 0 && line.Tag == "INDI" {
			individualCount++
		}
		return nil
	}
	for scanner.Scan() {
		lineText := scanner.Text()
		line, err := p.ParseLine(lineText)
		if err != nil {
			log.Fatalf("parse line: %v", err)
		}
		if err := handleLine(line); err != nil {
			log.Fatalf("handle line: %v", err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("scan: %v", err)
	}

	fmt.Printf("Individuals: %d\n", individualCount)
}
