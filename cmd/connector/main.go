package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jaedle/concourse-claude-connector/internal/concourse"
	"github.com/jaedle/concourse-claude-connector/internal/server"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if len(os.Args) > 1 && os.Args[1] == "health" {
		os.Exit(health(port))
	}

	config := server.Config{
		Concourse: concourse.Config{
			URL:      requireEnv("CONCOURSE_URL"),
			Username: requireEnv("CONCOURSE_USERNAME"),
			Password: requireEnv("CONCOURSE_PASSWORD"),
		},
		Version: version,
	}

	slog.Info("starting server", "port", port, "version", version)
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(config).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func requireEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		slog.Error("missing required environment variable", "name", name)
		os.Exit(1)
	}
	return value
}

func health(port string) int {
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(fmt.Sprintf("http://localhost:%s/health", port))
	if err != nil {
		return 1
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
