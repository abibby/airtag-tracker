package main

import (
	"log"
	"os"

	"github.com/abibby/airtag-tracker/process"
)

func main() {
	err := process.Handle(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
}
