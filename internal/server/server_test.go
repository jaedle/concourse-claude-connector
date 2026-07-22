package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jaedle/concourse-claude-connector/internal/server"
)

var _ = Describe("Server", func() {
	var testServer *httptest.Server

	BeforeEach(func() {
		testServer = httptest.NewServer(server.New().Handler())
		DeferCleanup(testServer.Close)
	})

	It("reports health", func() {
		response, err := http.Get(testServer.URL + "/health")

		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		body, err := io.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchJSON(`{"status":"ok"}`))
	})
})
