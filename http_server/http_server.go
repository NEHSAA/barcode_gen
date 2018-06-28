package http_server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func NewServer() Server {
	s := Server{
		Router: mux.NewRouter(),
	}

	s.Router.HandleFunc("/", handle)

	return s
}
