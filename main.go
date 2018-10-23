package main

import (
	"log"

	"github.com/NEHSAA/barcode_gen/app"
)

func main() {
	app := app.NewApp()
	log.Fatal(app.Run())
}
