// Command mcpsnapshot prints the advertised mcp tools as a snapshot for mcpgrade.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jaedle/concourse-claude-connector/internal/concourse"
	"github.com/jaedle/concourse-claude-connector/internal/mcp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	server := mcp.NewServer(concourse.NewClient(concourse.Config{}), "snapshot")
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		return fmt.Errorf("connect server: %w", err)
	}

	client := sdk.NewClient(&sdk.Implementation{Name: "mcpsnapshot", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		return fmt.Errorf("connect client: %w", err)
	}
	defer func() { _ = session.Close() }()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(map[string]any{
		"serverName": session.InitializeResult().ServerInfo.Name,
		"tools":      tools.Tools,
	})
}
