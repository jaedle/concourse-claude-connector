//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jaedle/concourse-claude-connector/internal/concourse/concoursetest"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = Describe("Connector", Ordered, func() {
	var fake *concoursetest.Fake
	var baseURL string
	project := fmt.Sprintf("concourse-claude-connector-e2e-%d", os.Getpid())

	compose := func(args ...string) string {
		command := exec.Command("docker", append([]string{"compose", "-p", project, "-f", "docker-compose.yaml"}, args...)...)
		command.Env = append(os.Environ(), "CONCOURSE_PORT="+fake.Port())
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		return string(output)
	}

	BeforeAll(func() {
		var err error
		fake, err = concoursetest.NewListening("0.0.0.0:0", "e2e-user", "e2e-password")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(fake.Close)

		DeferCleanup(func() { compose("down", "-v", "--rmi", "local") })
		compose("up", "-d", "--build")

		port := strings.TrimSpace(compose("port", "connector", "8080"))
		baseURL = "http://" + port

		Eventually(func() error {
			response, err := http.Get(baseURL + "/health")
			if err != nil {
				return err
			}
			defer func() { _ = response.Body.Close() }()
			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("health returned status %d", response.StatusCode)
			}
			return nil
		}).WithTimeout(30 * time.Second).WithPolling(500 * time.Millisecond).Should(Succeed())
	})

	It("lists pipelines via mcp", func() {
		fake.SetPipelines([]map[string]any{
			{"name": "deploy", "team_name": "main", "paused": false, "public": true, "archived": false},
		})

		client := sdk.NewClient(&sdk.Implementation{Name: "e2e-client", Version: "test"}, nil)
		session, err := client.Connect(
			context.Background(),
			&sdk.StreamableClientTransport{Endpoint: baseURL + "/mcp"},
			nil,
		)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = session.Close() }()

		result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "list_pipelines"})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.IsError).To(BeFalse())
		structured, err := json.Marshal(result.StructuredContent)
		Expect(err).NotTo(HaveOccurred())
		Expect(structured).To(MatchJSON(`{
			"pipelines": [
				{"name": "deploy", "team": "main", "paused": false, "public": true, "archived": false}
			]
		}`))
	})
})
