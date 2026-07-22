package mcp

import (
	"context"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jaedle/concourse-claude-connector/internal/concourse"
)

type Pipeline struct {
	Name     string `json:"name"`
	Team     string `json:"team"`
	Paused   bool   `json:"paused"`
	Public   bool   `json:"public"`
	Archived bool   `json:"archived"`
}

type ListPipelinesOutput struct {
	Pipelines []Pipeline `json:"pipelines"`
}

func NewHandler(client *concourse.Client, version string) http.Handler {
	server := sdk.NewServer(&sdk.Implementation{
		Name:    "concourse-claude-connector",
		Title:   "Concourse CI",
		Version: version,
	}, nil)

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_pipelines",
		Description: "List all Concourse pipelines visible to the configured user, across all teams.",
	}, listPipelines(client))

	return sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server {
		return server
	}, &sdk.StreamableHTTPOptions{Stateless: true})
}

func listPipelines(client *concourse.Client) func(context.Context, *sdk.CallToolRequest, struct{}) (*sdk.CallToolResult, ListPipelinesOutput, error) {
	return func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, ListPipelinesOutput, error) {
		pipelines, err := client.Pipelines(ctx)
		if err != nil {
			return nil, ListPipelinesOutput{}, err
		}

		output := ListPipelinesOutput{Pipelines: []Pipeline{}}
		for _, pipeline := range pipelines {
			output.Pipelines = append(output.Pipelines, Pipeline{
				Name:     pipeline.Name,
				Team:     pipeline.Team,
				Paused:   pipeline.Paused,
				Public:   pipeline.Public,
				Archived: pipeline.Archived,
			})
		}
		return nil, output, nil
	}
}
