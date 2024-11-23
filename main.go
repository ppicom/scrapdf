package main

import (
	"log"

	"github.com/ppicom/scrapedf/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
