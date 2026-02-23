package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
)

func TestNewMCPServer(t *testing.T) {
	s := NewMCPServer("name", "1.0.0")
	if s == nil || s.server == nil {
		t.Fatalf("expected MCP server to be initialized")
	}
}

func TestAddToolValidation(t *testing.T) {
	s := NewMCPServer("name", "1.0.0")
	err := s.AddTool(ToolDefinition{}, func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		return ResultText("ok"), nil
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestAddToolSuccess(t *testing.T) {
	s := NewMCPServer("name", "1.0.0")
	def := ToolDefinition{
		Name:        "echo",
		Description: "Echo tool",
		Schema: JSONSchema{
			Type: "object",
		},
	}
	err := s.AddTool(def, func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		return ResultText("ok"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddToolFromJSON(t *testing.T) {
	s := NewMCPServer("name", "1.0.0")
	valid := []byte(`{"name":"search","description":"Search","schema":{"type":"object"}}`)
	if err := s.AddToolFromJSON(valid, func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		return ResultText("ok"), nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.AddToolFromJSON([]byte(`{"name":}`), func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		return ResultText("ok"), nil
	}); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestAddToolFromFile(t *testing.T) {
	s := NewMCPServer("name", "1.0.0")
	dir := t.TempDir()
	path := filepath.Join(dir, "tool.json")
	data := []byte(`{"name":"search","description":"Search","schema":{"type":"object"}}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if err := s.AddToolFromFile(path, func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		return ResultText("ok"), nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartValidation(t *testing.T) {
	ctx := context.Background()
	s := &MCPServer{}
	if err := s.Start(ctx, ":0"); err == nil {
		t.Fatalf("expected error for nil server")
	}

	s = NewMCPServer("name", "1.0.0")
	if err := s.Start(ctx, " "); err == nil {
		t.Fatalf("expected error for empty address")
	}
}

func TestStartStdioValidation(t *testing.T) {
	ctx := context.Background()
	s := &MCPServer{}
	if err := s.StartStdio(ctx); err == nil {
		t.Fatalf("expected error for nil server")
	}
}

func TestShutdownNoop(t *testing.T) {
	ctx := context.Background()
	s := NewMCPServer("name", "1.0.0")
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWrapHandler(t *testing.T) {
	h := wrapHandler(func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		return nil, errors.New("boom")
	})
	res, err := h(context.Background(), mcpproto.CallToolRequest{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatalf("expected error result")
	}
	text := textContent(t, res)
	if !strings.Contains(text.Text, "tool execution error") {
		t.Fatalf("unexpected error text: %q", text.Text)
	}

	h = wrapHandler(func(ctx context.Context, req mcpproto.CallToolRequest) (*mcpproto.CallToolResult, error) {
		return ResultText("ok"), nil
	})
	res, err = h(context.Background(), mcpproto.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text = textContent(t, res)
	if text.Text != "ok" {
		t.Fatalf("unexpected text: %q", text.Text)
	}
}
