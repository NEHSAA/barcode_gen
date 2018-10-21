package app

import (
	"log"
	"net/http"
	"time"

	logging "github.com/NEHSAA/barcode_gen/common/log"
	"github.com/NEHSAA/barcode_gen/http_server"
)

type App struct {
	server *http_server.Server
}

func NewApp() *App {
	return &App{
		server: http_server.NewServer(),
	}
}

func (a *App) Run() {
	logging.InitLogrus()

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.server.Router, // Pass our instance of gorilla/mux in.
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
