package server

import (
	"encoding/json"
	"fmt"
)

// JSONSchemaProperty describes a single property inside a JSON Schema object.
type JSONSchemaProperty struct {
	Type        string              `json:"type"`
	Description string              `json:"description,omitempty"`
	Enum        []string            `json:"enum,omitempty"`
	Items       *JSONSchemaProperty `json:"items,omitempty"` // for type=array
	Default     any                 `json:"default,omitempty"`
}

// JSONSchema is the input schema attached to an MCP tool.
// It follows the JSON Schema draft-07 "object" convention.
type JSONSchema struct {
	Type       string                        `json:"type"`
	Properties map[string]JSONSchemaProperty `json:"properties,omitempty"`
	Required   []string                      `json:"required,omitempty"`
}

// ToolDefinition is the full description of a tool that will be registered
// in the MCP server. It can be built in code or loaded from a JSON file.
type ToolDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Schema      JSONSchema `json:"schema"`
}

// Validate checks that the minimum required fields are present.
func (d ToolDefinition) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if d.Description == "" {
		return fmt.Errorf("tool %q: description cannot be empty", d.Name)
	}
	if d.Schema.Type == "" {
		return fmt.Errorf("tool %q: schema.type cannot be empty", d.Name)
	}
	return nil
}

// SchemaBytes returns the JSON-encoded schema, ready to be passed to
// mcpproto.NewToolWithRawSchema.
func (d ToolDefinition) SchemaBytes() ([]byte, error) {
	return json.Marshal(d.Schema)
}

// ParseToolDefinition deserializes a ToolDefinition from raw JSON bytes.
func ParseToolDefinition(data []byte) (ToolDefinition, error) {
	var def ToolDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return ToolDefinition{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := def.Validate(); err != nil {
		return ToolDefinition{}, err
	}
	return def, nil
}
