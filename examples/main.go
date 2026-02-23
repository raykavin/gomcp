package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/raykavin/gomcp/server"
)

func main() {
	srv := mcpserver.NewMCPServer("example-agent", "1.0.0")

	registerTools(srv)

	// Transport
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Pick transport via env var:
	//   MCP_TRANSPORT=stdio  → stdin/stdout (for Claude Desktop, etc.)
	//   MCP_TRANSPORT=sse    → HTTP SSE on :8080 (default)
	switch os.Getenv("MCP_TRANSPORT") {
	case "stdio":
		log.Println("starting stdio transport")
		if err := srv.StartStdio(ctx); err != nil {
			log.Fatalf("stdio: %v", err)
		}
	default:
		addr := envOrDefault("MCP_ADDR", ":8080")
		log.Printf("starting SSE transport on %s", addr)
		if err := srv.Start(ctx, addr); err != nil {
			log.Fatalf("sse: %v", err)
		}
	}
}

func registerTools(srv *mcpserver.MCPServer) {
	// AddToolFunc: inline definition with explicit fields.
	mustAdd("AddToolFunc", srv.AddToolFunc(
		"echo",
		"Returns the input message back to the caller",
		mcpserver.JSONSchema{
			Type: "object",
			Properties: map[string]mcpserver.JSONSchemaProperty{
				"message": {
					Type:        "string",
					Description: "The message to echo",
				},
			},
			Required: []string{"message"},
		},
		func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
			msg, errResult := mcpserver.ArgStringRequired(req, "message")
			if errResult != nil {
				return errResult, nil
			}
			return mcpserver.ResultText("echo: " + msg), nil
		},
	))

	// AddTool: definition via ToolDefinition.
	def := mcpserver.ToolDefinition{
		Name:        "add",
		Description: "Adds two numbers and returns the result",
		Schema: mcpserver.JSONSchema{
			Type: "object",
			Properties: map[string]mcpserver.JSONSchemaProperty{
				"a": {Type: "number", Description: "First operand"},
				"b": {Type: "number", Description: "Second operand"},
			},
			Required: []string{"a", "b"},
		},
	}
	mustAdd("AddTool", srv.AddTool(def,
		func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
			a, ok := mcpserver.ArgFloat(req, "a")
			if !ok {
				return mcpserver.ResultErrorMsg("missing argument: a"), nil
			}
			b, ok := mcpserver.ArgFloat(req, "b")
			if !ok {
				return mcpserver.ResultErrorMsg("missing argument: b"), nil
			}
			return mcpserver.ResultJSON(map[string]any{"result": a + b}), nil
		}))

	// AddToolFromFile: load tool definition from JSON on disk.
	mustAdd("AddToolFromFile", srv.AddToolFromFile("tools/search.json",
		func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
			query, errResult := mcpserver.ArgStringRequired(req, "query")
			if errResult != nil {
				return errResult, nil
			}
			return mcpserver.ResultText(fmt.Sprintf("results for: %q", query)), nil
		}))
}

func mustAdd(label string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", label, err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
