# concourse-claude-connector

MCP server exposing [Concourse CI](https://concourse-ci.org/) to Claude.

## Tools

- `list_pipelines` — all pipelines visible to the configured user, across teams

## Configuration

Environment variables:

| Variable             | Default | Description                        |
|----------------------|---------|------------------------------------|
| `PORT`               | `8080`  | HTTP listen port                   |
| `CONCOURSE_URL`      | —       | Concourse base URL                 |
| `CONCOURSE_USERNAME` | —       | Concourse local user               |
| `CONCOURSE_PASSWORD` | —       | Concourse local user password      |

## Development

```sh
task ci
```
