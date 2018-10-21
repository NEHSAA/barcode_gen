package http_server

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ErrorGAuthTokenDoesNotExist = fmt.Errorf("oauth token does not exist")
	ErrorGAuthTokenExpired      = fmt.Errorf("oauth token expired")
)

func getAuthCconfig(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		RedirectURL:  getFullUrl(r, "/auth/callback"),
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

func getLoginURL(r *http.Request, state string) string {
	return getAuthCconfig(r).AuthCodeURL(state)
}

func (s *Server) handleAuthOK(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	session := s.getSession(r, "gauth")
	state := randToken()

	session.Values["state"] = state
	session.Save(r, w)
	url := getLoginURL(r, state)
	// w.Write([]byte("<html><title>Golang Google</title> <body> <a href='" + url + "'>Login with Google!</a> </body></html>"))

	query := r.URL.Query()
	redirect := query.Get("redirect")

	if redirect != "" {
		s.redirectToGAuthLogin(w, r, redirect)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	session := s.getSession(r, "gauth")
	var err error

	delete(session.Values, "token")

	query := r.URL.Query()
	redirect := query.Get("redirect")

	orig_loc, _ := session.Values["original_location"]
	if orig_loc != nil {
		delete(session.Values, "original_location")
		err = session.Save(r, w)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("save session error: %s", err)))
			return
		}
	}

	if redirect != "" || orig_loc != nil {
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
			return
		}
		http.Redirect(w, r, orig_loc.(string), http.StatusTemporaryRedirect)
		return
	}

	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("save session error: %s", err)))
		return
	}
	http.Redirect(w, r, "/auth/ok", http.StatusTemporaryRedirect)
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Handle the exchange code to initiate a transport.
	session := s.getSession(r, "gauth")
	vars := mux.Vars(r)
	retrievedState := session.Values["state"]
	if retrievedState != vars["state"] {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Invalid session state: %s\n", retrievedState)))
		w.Write([]byte("<p><a href=\"/auth/login\">retry login</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}

	tok, err := getAuthCconfig(r).Exchange(ctx, vars["code"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("<html><body>"))
		w.Write([]byte(fmt.Sprintf("Exchange error: %s\n", err)))
		w.Write([]byte("<p><a href=\"/auth/login\">retry login</a></p>"))
		w.Write([]byte("</html></body>"))
		return
	}
	session.Values["token"] = tok
	// session.Save(r, w)

	orig_loc, _ := session.Values["original_location"]
	if orig_loc != nil {
		delete(session.Values, "original_location")
		err = session.Save(r, w)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("<html><body>"))
			w.Write([]byte(fmt.Sprintf("save session error: %s\n", err)))
			w.Write([]byte("<p><a href=\"/auth/login\">retry login</a></p>"))
			w.Write([]byte("</html></body>"))
			return
		}
		http.Redirect(w, r, orig_loc.(string), http.StatusTemporaryRedirect)
		return
	}
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("save session error: %s", err)))
		return
	}
	http.Redirect(w, r, "/auth/ok", http.StatusTemporaryRedirect)
}

func (s *Server) getGoogleClient(ctx context.Context, r *http.Request) (*http.Client, error) {
	session := s.getSession(r, "gauth")
	val := session.Values["token"]
	var token = &oauth2.Token{}
	var ok bool
	if token, ok = val.(*oauth2.Token); !ok {
		return nil, ErrorGAuthTokenDoesNotExist
	}
	if time.Now().After(token.Expiry) {
		return nil, ErrorGAuthTokenExpired
	}

	return getAuthCconfig(r).Client(ctx, token), nil
}

func (s *Server) redirectToGAuthLogin(w http.ResponseWriter, r *http.Request, originalLocation string) {
	session, vals := s.getSessionStore(r, "gauth")
	(*vals)["original_location"] = originalLocation
	session.Save(r, w)

	http.Redirect(w, r, getFullUrl(r, "/auth/login"), http.StatusTemporaryRedirect)
}

func (s *Server) mountGAuthEndpoints() {
	gob.Register(&oauth2.Token{})

	s.Router.HandleFunc("/auth/login", s.handleLogin)
	s.Router.HandleFunc("/auth/logout", s.handleLogout)
	s.Router.HandleFunc("/auth/ok", s.handleAuthOK)
	s.Router.Path("/auth/callback").Queries("state", "{state}", "code", "{code}").HandlerFunc(s.handleAuth)
}
