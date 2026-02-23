package server

import (
	"encoding/json"
	"testing"
)

func TestToolDefinitionValidate(t *testing.T) {
	def := ToolDefinition{}
	if err := def.Validate(); err == nil {
		t.Fatalf("expected error for empty definition")
	}

	def = ToolDefinition{Name: "tool"}
	if err := def.Validate(); err == nil {
		t.Fatalf("expected error for missing description")
	}

	def = ToolDefinition{Name: "tool", Description: "desc"}
	if err := def.Validate(); err == nil {
		t.Fatalf("expected error for missing schema type")
	}

	def = ToolDefinition{
		Name:        "tool",
		Description: "desc",
		Schema:      JSONSchema{Type: "object"},
	}
	if err := def.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchemaBytes(t *testing.T) {
	def := ToolDefinition{
		Name:        "tool",
		Description: "desc",
		Schema: JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchemaProperty{
				"q": {Type: "string"},
			},
			Required: []string{"q"},
		},
	}

	raw, err := def.SchemaBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got JSONSchema
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("expected valid JSON schema, got error: %v", err)
	}
	if got.Type != "object" || got.Properties["q"].Type != "string" {
		t.Fatalf("unexpected schema: %#v", got)
	}
}

func TestParseToolDefinition(t *testing.T) {
	valid := []byte(`{"name":"search","description":"Search","schema":{"type":"object"}}`)
	def, err := ParseToolDefinition(valid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "search" || def.Schema.Type != "object" {
		t.Fatalf("unexpected definition: %#v", def)
	}

	if _, err := ParseToolDefinition([]byte(`{"name":}`)); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
	if _, err := ParseToolDefinition([]byte(`{"name":"x","schema":{"type":"object"}}`)); err == nil {
		t.Fatalf("expected error for missing description")
	}
}
