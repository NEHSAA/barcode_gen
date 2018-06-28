//+build !appengine

package app

import (
	"net/http"

	"github.com/NEHSAA/barcode_gen/http_server"
)

func Init() {
	server := http_server.NewServer()
	http.Handle("/", server.Router)
}
