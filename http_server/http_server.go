package http_server

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"common/log"
	"pdf"

	"github.com/gorilla/mux"

	"google.golang.org/appengine"
)

type Server struct {
	Router *mux.Router
}

func handle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	logger := log.GetLogger(ctx)
	logger.Debugf("111")
	vars := mux.Vars(r)
	fmt.Fprintln(w, "Hello, world! "+vars["path"])
}

func handleMakePdf(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	logger := log.GetLogger(ctx)

	vars := mux.Vars(r)
	logger.Debugf("making pdf based on %v", vars)
	data := []pdf.IdBarcodeData{
		{vars["name"] + pdf.TextSeparator + vars["member"],
			vars["barcode"]},
	}
	pdf, err := pdf.GetIdBarcodePdf(ctx, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	content := bytes.NewReader(pdf)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename='nehsaa_barcode.pdff")
	http.ServeContent(w, r, "pdf", time.Now(), content)
}

func NewServer() Server {
	s := Server{
		Router: mux.NewRouter(),
	}

	s.Router.HandleFunc("/make/{name}/{member}/{barcode}", handleMakePdf)
	s.Router.HandleFunc("/{path}", handle)

	return s
}
