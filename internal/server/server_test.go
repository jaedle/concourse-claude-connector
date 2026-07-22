package server_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jaedle/concourse-claude-connector/internal/concourse"
	"github.com/jaedle/concourse-claude-connector/internal/concourse/concoursetest"
	"github.com/jaedle/concourse-claude-connector/internal/server"
)

var _ = Describe("Server", func() {
	var fake *concoursetest.Fake
	var testServer *httptest.Server

	BeforeEach(func() {
		fake = concoursetest.New("a-user", "a-password")
		DeferCleanup(fake.Close)

		testServer = httptest.NewServer(server.New(server.Config{
			Concourse: concourse.Config{
				URL:      fake.URL(),
				Username: "a-user",
				Password: "a-password",
			},
			Version: "test",
		}).Handler())
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

	Describe("mcp", func() {
		var session *sdk.ClientSession

		BeforeEach(func() {
			client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "test"}, nil)
			var err error
			session, err = client.Connect(
				context.Background(),
				&sdk.StreamableClientTransport{Endpoint: testServer.URL + "/mcp"},
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { _ = session.Close() })
		})

		It("exposes the list_pipelines tool", func() {
			tools, err := session.ListTools(context.Background(), nil)

			Expect(err).NotTo(HaveOccurred())
			names := []string{}
			for _, tool := range tools.Tools {
				names = append(names, tool.Name)
			}
			Expect(names).To(Equal([]string{"list_pipelines"}))
		})

		It("lists pipelines", func() {
			fake.SetPipelines([]map[string]any{
				{"name": "deploy", "team_name": "main", "paused": true, "public": false, "archived": false},
			})

			result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "list_pipelines"})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsError).To(BeFalse())
			structured, err := json.Marshal(result.StructuredContent)
			Expect(err).NotTo(HaveOccurred())
			Expect(structured).To(MatchJSON(`{
				"pipelines": [
					{"name": "deploy", "team": "main", "paused": true, "public": false, "archived": false}
				]
			}`))
		})

		It("reports tool errors", func() {
			fake.Close()

			result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "list_pipelines"})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsError).To(BeTrue())
		})
	})
})
