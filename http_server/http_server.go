package http_server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/NEHSAA/barcode_gen/common/log"
	"github.com/NEHSAA/barcode_gen/pdf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type Server struct {
	Router *mux.Router
	logger log.Logger
	store  *sessions.CookieStore
}

func getFullUrl(r *http.Request, path string) string {
	rurl := &url.URL{
		Scheme: r.URL.Scheme,
		Host:   r.Host,
		Path:   path,
	}
	if rurl.Scheme == "" {
		rurl.Scheme = "http"
	}
	return rurl.String()
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, fmt.Sprintf("Hello, world!"))
}

func (s *Server) handleMakePdf(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogrusLogger("handler")

	vars := mux.Vars(r)
	logger.Debugf("making pdf based on %v", vars)
	data := []pdf.IdBarcodeData{
		{vars["name"] + pdf.TextSeparator + vars["member"],
			vars["barcode"]},
	}
	pdf, err := pdf.GetIdBarcodePdf(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	content := bytes.NewReader(pdf)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename='nehsaa_barcode.pdf")
	http.ServeContent(w, r, "pdf", time.Now(), content)
}

func (s *Server) handleMakeMultiPdf(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogrusLogger("handler")

	vars := r.URL.Query()
	logger.Debugf("making pdf based on %v", vars)

	if len(vars) == 0 {
		fmt.Fprintln(w, "no data to show")
		return
	}

	names, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "key \"name\" is missing")
		return
	}

	types, ok := vars["type"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "key \"type\" is missing")
		return
	}

	barcodes, ok := vars["barcode"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "key \"barcode\" is missing")
		return
	}

	if (len(names) != len(types)) || (len(names) != len(barcodes)) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "provided number of fields are inconsistent")
		return
	}

	data := make([]pdf.IdBarcodeData, len(names))
	for i, name := range names {
		data[i].Text = name + pdf.TextSeparator + types[i]
		data[i].BarcodeContent = barcodes[i]
		//  = {name + pdf.TextSeparator + types[i],	barcodes[i]}
	}

	pdf, err := pdf.GetIdBarcodePdf(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	content := bytes.NewReader(pdf)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename='nehsaa_barcode.pdf")
	http.ServeContent(w, r, "pdf", time.Now(), content)
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (s *Server) getSession(r *http.Request, name string) *sessions.Session {
	session, _ := s.store.Get(r, name)
	return session
}

func (s *Server) getSessionStore(r *http.Request, name string) (*sessions.Session, *map[interface{}]interface{}) {
	session := s.getSession(r, name)
	return session, &session.Values
}

func NewServer() *Server {
	s := &Server{
		Router: mux.NewRouter(),
		logger: log.GetLogrusLogger("http_server"),
		store:  sessions.NewCookieStore([]byte(randToken())),
	}

	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.logger.Infof("%v", r.RequestURI)
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	}
	s.Router.Use(loggingMiddleware)

	s.Router.HandleFunc("/makepdf/multi", s.handleMakeMultiPdf)
	s.Router.HandleFunc("/makepdf/{name}/{member}/{barcode}", s.handleMakePdf)

	s.Router.HandleFunc("/", s.handleRoot)
	s.mountGAuthEndpoints()
	s.mountMemberSpreadsheetEndpoints()

	return s
}
