//+build appengine

package app

import (
	"http_server"
	"net/http"
)

func Init() {
	server := http_server.NewServer()
	http.Handle("/", server.Router)
}
