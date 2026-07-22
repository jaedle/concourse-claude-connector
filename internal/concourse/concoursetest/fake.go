// Package concoursetest provides a fake Concourse for tests.
package concoursetest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

type LoginRequest struct {
	ClientID     string
	ClientSecret string
	GrantType    string
	Scope        string
}

type Fake struct {
	server *httptest.Server

	mutex        sync.Mutex
	username     string
	password     string
	pipelines    []map[string]any
	logins       int
	currentToken string
	lastLogin    LoginRequest
}

func New(username, password string) *Fake {
	fake := &Fake{username: username, password: password}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /sky/issuer/token", fake.handleToken)
	mux.HandleFunc("GET /api/v1/pipelines", fake.handlePipelines)
	fake.server = httptest.NewServer(mux)

	return fake
}

func (f *Fake) handleToken(w http.ResponseWriter, r *http.Request) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	clientID, clientSecret, _ := r.BasicAuth()
	_ = r.ParseForm()
	f.lastLogin = LoginRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    r.PostFormValue("grant_type"),
		Scope:        r.PostFormValue("scope"),
	}

	if clientID != "fly" || clientSecret != "Zmx5" ||
		r.PostFormValue("username") != f.username ||
		r.PostFormValue("password") != f.password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f.logins++
	f.currentToken = fmt.Sprintf("token-%d", f.logins)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": f.currentToken,
		"token_type":   "bearer",
	})
}

func (f *Fake) handlePipelines(w http.ResponseWriter, r *http.Request) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.currentToken == "" || r.Header.Get("Authorization") != "Bearer "+f.currentToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(f.pipelines)
}

func (f *Fake) SetPipelines(pipelines []map[string]any) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.pipelines = pipelines
}

func (f *Fake) InvalidateToken() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.currentToken = ""
}

func (f *Fake) LoginCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.logins
}

func (f *Fake) LastLogin() LoginRequest {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.lastLogin
}

func (f *Fake) URL() string {
	return f.server.URL
}

func (f *Fake) Close() {
	f.server.Close()
}
