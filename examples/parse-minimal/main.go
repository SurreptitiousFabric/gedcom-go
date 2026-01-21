package main

import (
	"log"
	"os"

	"github.com/cacack/gedcom-go/decoder"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := decoder.Decode(f); err != nil {
		log.Fatal(err)
	}
}
