# Go MCP Server 

A clean Go library for building [Model Context Protocol (MCP)](https://modelcontextprotocol.io) servers on top of [mcp-go](https://github.com/mark3labs/mcp-go).

## Features
- Three ways to register tools: inline struct, function helper, or JSON file on disk
- Built-in argument extraction helpers (`ArgString`, `ArgFloat`, `ArgBool`, …)
- Result builders (`ResultText`, `ResultJSON`, `ResultError`, …)
- Both **SSE (HTTP)** and **Stdio** transports — pick at runtime

---

## Installation

```bash
go get github.com/raykavin/gomcp
```

---

## Quick start

```go
package main

import (
	"context"

	mcpserver "github.com/raykavin/gomcp/server"
)

func main() {

	srv := mcpserver.NewMCPServer("my-agent", "1.0.0")

	// Register a tool
	srv.AddToolFunc(
		"echo",
		"Returns the input message",
		mcpserver.JSONSchema{
			Type: "object",
			Required: []string{"message"},
			Properties: map[string]mcpserver.JSONSchemaProperty{
				"message": {Type: "string", Description: "Message to echo"},
			},
		},
		func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
			msg, errResult := mcpserver.ArgStringRequired(req, "message")
			if errResult != nil {
				return errResult, nil
			}
			return mcpserver.ResultText("echo: " + msg), nil
		},
	)

	// Start SSE transport
	srv.Start(ctx, ":8080")
}

```
---

## Tool registration methods

### 1. `AddToolFunc`:  inline, all fields explicit

Best for tools defined entirely in Go code.

```go
err := srv.AddToolFunc(
    "search",                   // name
    "Search the web",           // description
    mcpserver.JSONSchema{...},  // schema
    myHandler,
)
```

### 2. `AddTool`: using a `ToolDefinition` struct

Useful when building definitions programmatically or receiving them from external sources.

```go
def := mcpserver.ToolDefinition{
    Name:        "add",
    Description: "Adds two numbers",
    Schema:      mcpserver.JSONSchema{...},
}
err := srv.AddTool(def, myHandler)
```

### 3. `AddToolFromFile`: load from a JSON file on disk

Keeps tool metadata outside Go code. Ideal for config-driven agents or tools shared across projects.

```go
err := srv.AddToolFromFile("tools/search.json", myHandler)
```

**JSON file format:**

```json
{
  "name": "search",
  "description": "Search the web and return relevant results",
  "schema": {
    "type": "object",
    "properties": {
      "query": { "type": "string", "description": "The search query" },
      "max_results": { "type": "number", "description": "Max results" }
    },
    "required": ["query"]
  }
}
```

### 4. `AddToolFromJSON`: load from raw `[]byte`

For tools fetched from a database, remote config, etc.

```go
data := []byte(`{...}`)
err := srv.AddToolFromJSON(data, myHandler)
```

---

## Argument helpers

Inside a handler, use the helpers to safely extract typed arguments:

```go
// String (returns error result if missing)
msg, errResult := mcpserver.ArgStringRequired(req, "message")
if errResult != nil {
    return errResult, nil
}

// Optional string
if q, ok := mcpserver.ArgString(req, "query"); ok { ... }

// Number
if n, ok := mcpserver.ArgFloat(req, "count"); ok { ... }

// Bool
if verbose, ok := mcpserver.ArgBool(req, "verbose"); ok { ... }
```

---

## Result helpers

```go
mcpserver.ResultText("plain string response")
mcpserver.ResultJSON(map[string]any{"key": "value"})
mcpserver.ResultError("something broke", err)
mcpserver.ResultErrorMsg("missing argument: query")
```

---

## Transports

### SSE (HTTP): default

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

if err := srv.Start(ctx, ":8080"); err != nil {
    log.Fatal(err)
}
```

Clients connect via `http://localhost:8080/sse`.

### Stdio: For Claude Desktop and local agents

```go
if err := srv.StartStdio(ctx); err != nil {
    log.Fatal(err)
}
```

### Graceful shutdown (SSE only)

```go
srv.Shutdown(ctx)
```

---

## Project structure

```
github.com/raykavin/gomcp/
├── server.go     # MCPServer, Start, StartStdio, Shutdown
├── tool.go       # ToolDefinition, JSONSchema, ParseToolDefinition
├── helpers.go    # ResultText/JSON/Error, ArgString/Float/Bool helpers
├── go.mod
└── example/
    ├── main.go
    └── tools/
        └── search.json
```

---

## Tests

```bash
go test ./server
```

---

## Contributing

Contributions to Go MCP Server are welcome! Here are some ways you can help improve the project:

- **Report bugs and suggest features** by opening issues on GitHub
- **Submit pull requests** with bug fixes or new features
- **Improve documentation** to help other users and developers
- **Share your custom strategies** with the community

## License

Go MCP Server is distributed under the **MIT License**.  
For complete license terms and conditions, see the [LICENSE](LICENSE.md) file in the repository.

---

## Contact

For support, collaboration, or questions about Go MCP Server:

**Email**: [raykavin.meireles@gmail.com](mailto:raykavin.meireles@gmail.com)  
**GitHub**: [@raykavin](https://github.com/raykavin)  
