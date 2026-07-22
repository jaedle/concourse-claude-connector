package concourse_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jaedle/concourse-claude-connector/internal/concourse"
	"github.com/jaedle/concourse-claude-connector/internal/concourse/concoursetest"
)

var _ = Describe("Client", func() {
	var fake *concoursetest.Fake
	var client *concourse.Client

	BeforeEach(func() {
		fake = concoursetest.New("a-user", "a-password")
		DeferCleanup(fake.Close)

		client = concourse.NewClient(concourse.Config{
			URL:      fake.URL(),
			Username: "a-user",
			Password: "a-password",
		})
	})

	Describe("Pipelines", func() {
		It("lists pipelines across teams", func() {
			fake.SetPipelines([]map[string]any{
				{"name": "deploy", "team_name": "main", "paused": false, "public": true, "archived": false},
				{"name": "nightly", "team_name": "other", "paused": true, "public": false, "archived": true},
			})

			pipelines, err := client.Pipelines(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(pipelines).To(Equal([]concourse.Pipeline{
				{Name: "deploy", Team: "main", Paused: false, Public: true, Archived: false},
				{Name: "nightly", Team: "other", Paused: true, Public: false, Archived: true},
			}))
		})

		It("logs in with the fly client credentials", func() {
			_, err := client.Pipelines(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(fake.LastLogin()).To(Equal(concoursetest.LoginRequest{
				ClientID:     "fly",
				ClientSecret: "Zmx5",
				GrantType:    "password",
				Scope:        "openid profile email federated:id groups",
			}))
		})

		It("reuses the token across requests", func() {
			_, err := client.Pipelines(context.Background())
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Pipelines(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(fake.LoginCount()).To(Equal(1))
		})

		It("logs in again when the token is rejected", func() {
			_, err := client.Pipelines(context.Background())
			Expect(err).NotTo(HaveOccurred())
			fake.InvalidateToken()

			_, err = client.Pipelines(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(fake.LoginCount()).To(Equal(2))
		})

		It("fails on invalid credentials", func() {
			client = concourse.NewClient(concourse.Config{
				URL:      fake.URL(),
				Username: "a-user",
				Password: "wrong-password",
			})

			_, err := client.Pipelines(context.Background())

			Expect(err).To(MatchError(ContainSubstring("login failed")))
		})

		It("fails when concourse is unreachable", func() {
			client = concourse.NewClient(concourse.Config{
				URL:      "http://localhost:0",
				Username: "a-user",
				Password: "a-password",
			})

			_, err := client.Pipelines(context.Background())

			Expect(err).To(HaveOccurred())
		})
	})
})
