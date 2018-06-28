// +build appengine
package main

import (
	"app"

	"google.golang.org/appengine"
)

func main() {
	app.Init()
	appengine.Main()
}
