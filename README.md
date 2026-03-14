# hn-mcp

An MCP server that gives any MCP-compatible client (like Claude Code) the ability to browse, search, and read Hacker News using the official Firebase API and Algolia Search API, with no API keys or external dependencies required.

## Tools

| Tool | Description |
|------|-------------|
| `hn_front_page` | Fetch stories from any HN feed (top, new, best, ask, show, job) |
| `hn_search` | Search HN history with filters for date range, minimum points, and content type |
| `hn_comments` | Fetch a story's comment tree with configurable depth and limit |
| `hn_user` | Look up a user profile and optionally their recent submissions |

## Build

```
go build -o hn-mcp
```

## Configure

Add the server to your MCP client config. For Claude Code, add this to `.mcp.json`:

```json
{
  "mcpServers": {
    "hackernews": {
      "command": "/path/to/hn-mcp",
      "args": []
    }
  }
}
```

Replace `/path/to/hn-mcp` with the actual path to the built binary.
