// Package server provides a high-level wrapper around mcp-go,
// making it easy to build MCP servers with typed tool registration,
// JSON-file-based tool definitions, and both SSE and Stdio transports.
package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// ToolHandler is the function signature for tool handlers.
// req carries the full MCP CallToolRequest (name + arguments).
type ToolHandler func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error)

// MCPServer exposes AI-agent tools via the Model Context Protocol.
type MCPServer struct {
	server      *mcpserver.MCPServer
	sseServer   *mcpserver.SSEServer
	stdioServer *mcpserver.StdioServer
}

// NewMCPServer creates a new MCPServer with the given name and version.
//
//	srv := mcpserver.NewMCPServer("my-agent", "1.0.0")
func NewMCPServer(srvName, srvVersion string) *MCPServer {
	return &MCPServer{
		server: mcpserver.NewMCPServer(srvName, srvVersion),
	}
}

// AddTool registers a tool using an explicit ToolDefinition and a handler.
// This is the lowest-level registration method — all other Add* methods
// ultimately call this one.
func (s *MCPServer) AddTool(def ToolDefinition, handler ToolHandler) error {
	if err := def.Validate(); err != nil {
		return fmt.Errorf("invalid tool definition: %w", err)
	}

	schemaRaw, err := def.SchemaBytes()
	if err != nil {
		return fmt.Errorf("failed to marshal schema for tool %q: %w", def.Name, err)
	}

	mcpTool := mcpproto.NewToolWithRawSchema(def.Name, def.Description, schemaRaw)
	s.server.AddTool(mcpTool, wrapHandler(handler))
	return nil
}

// AddToolFunc registers a tool using individual fields plus a handler.
// schema must be a valid JSON Schema object (type/properties/required).
//
//	err := srv.AddToolFunc("search", "Search the web", schema, myHandler)
func (s *MCPServer) AddToolFunc(name, description string, schema JSONSchema, handler ToolHandler) error {
	return s.AddTool(ToolDefinition{
		Name:        name,
		Description: description,
		Schema:      schema,
	}, handler)
}

// AddToolFromFile loads a ToolDefinition from a JSON file on disk and
// registers it with the provided handler.
//
// The JSON file must match the ToolDefinition struct:
//
//	{
//	  "name":        "search",
//	  "description": "Search the web",
//	  "schema": {
//	    "type": "object",
//	    "properties": {
//	      "query": { "type": "string", "description": "Search query" }
//	    },
//	    "required": ["query"]
//	  }
//	}
func (s *MCPServer) AddToolFromFile(path string, handler ToolHandler) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read tool file %q: %w", path, err)
	}
	return s.AddToolFromJSON(data, handler)
}

// AddToolFromJSON parses a ToolDefinition from raw JSON bytes and registers
// the tool with the provided handler.
func (s *MCPServer) AddToolFromJSON(data []byte, handler ToolHandler) error {
	def, err := ParseToolDefinition(data)
	if err != nil {
		return fmt.Errorf("failed to parse tool definition: %w", err)
	}
	return s.AddTool(def, handler)
}

// Start starts the SSE (HTTP) transport on the given address (e.g. ":8080").
// It blocks until the context is cancelled or a fatal error occurs.
func (s *MCPServer) Start(ctx context.Context, addr string) error {
	if s.server == nil {
		return fmt.Errorf("mcp server not initialized")
	}
	if strings.TrimSpace(addr) == "" {
		return fmt.Errorf("invalid MCP address: empty string")
	}

	s.sseServer = mcpserver.NewSSEServer(s.server)

	go func() {
		<-ctx.Done()
		_ = s.sseServer.Shutdown(context.Background())
	}()

	return s.sseServer.Start(addr)
}

// StartStdio starts the Stdio transport, reading from os.Stdin and writing
// to os.Stdout. Blocks until ctx is cancelled or an error occurs.
func (s *MCPServer) StartStdio(ctx context.Context) error {
	if s.server == nil {
		return fmt.Errorf("mcp server not initialized")
	}

	s.stdioServer = mcpserver.NewStdioServer(s.server)
	return s.stdioServer.Listen(ctx, os.Stdin, os.Stdout)
}

// Shutdown gracefully stops the SSE server. No-op if SSE was never started.
func (s *MCPServer) Shutdown(ctx context.Context) error {
	if s.sseServer != nil {
		return s.sseServer.Shutdown(ctx)
	}
	return nil
}

// wrapHandler adapts a ToolHandler into the mcp-go callback signature,
// converting errors into MCP error results instead of propagating them.
func wrapHandler(h ToolHandler) func(context.Context, mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
	return func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		result, err := h(ctx, req)
		if err != nil {
			return mcpproto.NewToolResultErrorFromErr("tool execution error", err), nil
		}
		return result, nil
	}
}
