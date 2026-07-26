# AGENTS.md

## What this is

MCP server exposing Concourse CI to Claude. Go, deployed as a docker image
(`jaedle/concourse-claude-connector`).

Currently unauthenticated — OAuth 2.1 and public exposure via tunnel are not
implemented yet.

## Layout

- `cmd/connector` — entrypoint, env config, `health` subcommand
- `internal/server` — HTTP routing: `/mcp`, `/health`
- `internal/mcp` — MCP tools, Streamable HTTP handler (stateless)
- `internal/concourse` — Concourse API client (sky token login, token cache)
- `internal/concourse/concoursetest` — fake Concourse for tests
- `test/e2e` — compose-based end-to-end test (build tag `e2e`)
- `tools/mcpsnapshot` — prints the advertised tools; input for `mcpgrade`
- `build/package/Dockerfile` — multi-stage build to `scratch`
- `deployments/` — example deployment
- `ci/config.yaml` — consumed by `jaedle/pipeline-service`, which generates the
  Concourse pipeline; there is no pipeline yaml in this repo

## Workflow

- `task ci` must be green — that is the definition of done
- The MCP interface is linted by `mcpgrade` (`task test:mcp`) against a snapshot
  generated in-process; no server needs to run. Tool descriptions and schemas
  are the interface documentation — keep them meaningful.
  `.mcpgraderc.json` disables S004: it flags argument-less tools, which is a
  false positive here.
- Dependencies are vendored: after changing them run `task manage-dependencies`
- Commits: gitmoji + conventional commits (`✨ feat: ...`, `🔧 chore: ...`)
- Releases are automatic via semantic-release on `main`
- Toolchain versions come from `.mise.toml`; `mise install` provides them

## Conventions

- Configuration via environment variables only (see `README.md`)
- Logging: `log/slog`, JSON to stderr
- Tests: ginkgo/gomega, `_test.go` alongside the code, arrange/act/assert
  separated by blank lines
- No standalone golangci-lint config — defaults of golangci-lint v2
- Verify changes with a full `task ci` before committing.