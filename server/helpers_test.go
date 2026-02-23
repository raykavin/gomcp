package server

import (
	"encoding/json"
	"errors"
	"testing"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
)

func textContent(t *testing.T, result *mcpproto.CallToolResult) mcpproto.TextContent {
	t.Helper()
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	text, ok := result.Content[0].(mcpproto.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return text
}

func TestResultText(t *testing.T) {
	res := ResultText("hello")
	text := textContent(t, res)
	if text.Text != "hello" {
		t.Fatalf("expected text 'hello', got %q", text.Text)
	}
	if res.IsError {
		t.Fatalf("expected IsError=false")
	}
}

func TestResultJSONSuccess(t *testing.T) {
	input := map[string]any{"a": "b", "n": float64(2)}
	res := ResultJSON(input)
	text := textContent(t, res)

	var got map[string]any
	if err := json.Unmarshal([]byte(text.Text), &got); err != nil {
		t.Fatalf("expected valid JSON text, got error: %v", err)
	}
	if got["a"] != "b" || got["n"] != float64(2) {
		t.Fatalf("unexpected JSON content: %#v", got)
	}
	if res.StructuredContent == nil {
		t.Fatalf("expected structured content to be set")
	}
}

func TestResultJSONInvalid(t *testing.T) {
	res := ResultJSON(func() {})
	text := textContent(t, res)
	if text.Text != "{}" {
		t.Fatalf("expected fallback text '{}', got %q", text.Text)
	}
	if res.StructuredContent != nil {
		t.Fatalf("expected structured content to be nil")
	}
}

func TestResultError(t *testing.T) {
	err := errors.New("boom")
	res := ResultError("failed", err)
	text := textContent(t, res)
	if !res.IsError {
		t.Fatalf("expected IsError=true")
	}
	if text.Text == "" || text.Text == "failed" {
		t.Fatalf("expected error text to include details, got %q", text.Text)
	}
}

func TestResultErrorMsg(t *testing.T) {
	res := ResultErrorMsg("oops")
	text := textContent(t, res)
	if text.Text != "[error] oops" {
		t.Fatalf("expected error text '[error] oops', got %q", text.Text)
	}
	if res.IsError {
		t.Fatalf("expected IsError=false")
	}
}

func TestArgHelpers(t *testing.T) {
	req := mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{
			Arguments: map[string]any{
				"s": "str",
				"n": float64(3.14),
				"b": true,
			},
		},
	}

	if v, ok := ArgString(req, "s"); !ok || v != "str" {
		t.Fatalf("expected ArgString to return 'str' true, got %q %v", v, ok)
	}
	if v, ok := ArgFloat(req, "n"); !ok || v != 3.14 {
		t.Fatalf("expected ArgFloat to return 3.14 true, got %v %v", v, ok)
	}
	if v, ok := ArgBool(req, "b"); !ok || v != true {
		t.Fatalf("expected ArgBool to return true, got %v %v", v, ok)
	}
	if _, ok := ArgString(req, "missing"); ok {
		t.Fatalf("expected missing ArgString to return ok=false")
	}

	if v, errResult := ArgStringRequired(req, "missing"); errResult == nil || v != "" {
		t.Fatalf("expected missing ArgStringRequired to return error result")
	}
	_, errResult := ArgStringRequired(req, "missing")
	text := textContent(t, errResult)
	if text.Text != "[error] missing required argument: missing" {
		t.Fatalf("unexpected error text: %q", text.Text)
	}
}
