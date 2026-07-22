package server

import (
	"net/http"

	"github.com/jaedle/concourse-claude-connector/internal/concourse"
	"github.com/jaedle/concourse-claude-connector/internal/mcp"
)

type Config struct {
	Concourse concourse.Config
	Version   string
}

type Server struct {
	handler http.Handler
}

func New(config Config) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/mcp", mcp.NewHandler(concourse.NewClient(config.Concourse), config.Version))

	return &Server{handler: mux}
}

func (s *Server) Handler() http.Handler {
	return s.handler
}
