package server

import (
	"encoding/json"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
)

// ResultText returns a plain-text MCP tool result.
func ResultText(text string) *mcpproto.CallToolResult {
	return mcpproto.NewToolResultText(text)
}

// ResultJSON serializes v to JSON and returns a structured MCP tool result.
// Falls back to a text result containing the raw JSON string if marshalling
// fails (so handlers never need to handle the error themselves).
func ResultJSON(v any) *mcpproto.CallToolResult {
	raw, err := json.Marshal(v)
	if err != nil {
		return mcpproto.NewToolResultText("{}")
	}
	return mcpproto.NewToolResultStructured(v, string(raw))
}

// ResultError wraps an error message into an MCP error result.
func ResultError(msg string, err error) *mcpproto.CallToolResult {
	return mcpproto.NewToolResultErrorFromErr(msg, err)
}

// ResultErrorMsg wraps a plain string message into an MCP error result.
func ResultErrorMsg(msg string) *mcpproto.CallToolResult {
	return mcpproto.NewToolResultText("[error] " + msg)
}

// args safely casts req.Params.Arguments to map[string]any.
// Returns nil if the cast fails (e.g. Arguments is nil or a different type).
func args(req mcpproto.CallToolRequest) map[string]any {
	m, _ := req.Params.Arguments.(map[string]any)
	return m
}

// ArgString extracts a string argument from a CallToolRequest by key.
// Returns ("", false) if the key is absent or the value is not a string.
func ArgString(req mcpproto.CallToolRequest, key string) (string, bool) {
	v, ok := args(req)[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// ArgFloat extracts a float64 argument from a CallToolRequest by key.
func ArgFloat(req mcpproto.CallToolRequest, key string) (float64, bool) {
	v, ok := args(req)[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

// ArgBool extracts a bool argument from a CallToolRequest by key.
func ArgBool(req mcpproto.CallToolRequest, key string) (bool, bool) {
	v, ok := args(req)[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// ArgStringRequired extracts a required string argument.
// Returns ("", errResult) if absent or wrong type — intended for use inside
// handlers that want early returns on missing args.
func ArgStringRequired(req mcpproto.CallToolRequest, key string) (string, *mcpproto.CallToolResult) {
	v, ok := ArgString(req, key)
	if !ok {
		return "", ResultErrorMsg("missing required argument: " + key)
	}
	return v, nil
}
