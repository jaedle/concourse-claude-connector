package concourse_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

type fakeConcourse struct {
	server *httptest.Server

	mutex        sync.Mutex
	username     string
	password     string
	pipelines    []map[string]any
	logins       int
	currentToken string
	lastLogin    loginRequest
}

type loginRequest struct {
	clientID     string
	clientSecret string
	grantType    string
	scope        string
}

func newFakeConcourse(username, password string) *fakeConcourse {
	fake := &fakeConcourse{username: username, password: password}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /sky/issuer/token", fake.handleToken)
	mux.HandleFunc("GET /api/v1/pipelines", fake.handlePipelines)
	fake.server = httptest.NewServer(mux)

	return fake
}

func (f *fakeConcourse) handleToken(w http.ResponseWriter, r *http.Request) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	clientID, clientSecret, _ := r.BasicAuth()
	_ = r.ParseForm()
	f.lastLogin = loginRequest{
		clientID:     clientID,
		clientSecret: clientSecret,
		grantType:    r.PostFormValue("grant_type"),
		scope:        r.PostFormValue("scope"),
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

func (f *fakeConcourse) handlePipelines(w http.ResponseWriter, r *http.Request) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if r.Header.Get("Authorization") != "Bearer "+f.currentToken || f.currentToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(f.pipelines)
}

func (f *fakeConcourse) setPipelines(pipelines []map[string]any) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.pipelines = pipelines
}

func (f *fakeConcourse) invalidateToken() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.currentToken = ""
}

func (f *fakeConcourse) loginCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.logins
}

func (f *fakeConcourse) lastLoginRequest() loginRequest {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.lastLogin
}

func (f *fakeConcourse) url() string {
	return f.server.URL
}

func (f *fakeConcourse) close() {
	f.server.Close()
}
