package server

import (
	"net/http"
)

type Server struct {
	handler http.Handler
}

func New() *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return &Server{handler: mux}
}

func (s *Server) Handler() http.Handler {
	return s.handler
}
