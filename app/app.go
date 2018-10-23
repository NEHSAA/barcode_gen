package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/NEHSAA/barcode_gen/common/config"
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

func (a *App) Run() error {
	logging.InitLogrus()
	logger := logging.GetLogrusLogger("app")

	err := config.InitConfig()
	if err != nil {
		logger.Fatalf("error loading config: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		logger.Infof("Defaulting to port %s", port)
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.server.Router, // Pass our instance of gorilla/mux in.
	}

	return srv.ListenAndServe()
}
