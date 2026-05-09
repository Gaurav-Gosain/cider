package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Gaurav-Gosain/cider/pkg/fm"
)

// proxyTool implements fm.Tool and intercepts tool calls from the FM runtime
// to return them to the OpenAI API client instead of executing them.
type proxyTool struct {
	toolName string
	toolDesc string
	schema   *fm.GenerationSchema

	mu       sync.Mutex
	called   bool
	callArgs string
	cancel   context.CancelFunc
}

func (t *proxyTool) Name() string                     { return t.toolName }
func (t *proxyTool) Description() string               { return t.toolDesc }
func (t *proxyTool) ArgumentsSchema() *fm.GenerationSchema { return t.schema }

// Call is invoked by the FM runtime when the model wants to use this tool.
// Instead of executing the tool, we capture the arguments and cancel the
// generation context so the handler can return tool_calls to the client.
func (t *proxyTool) Call(args *fm.GeneratedContent) (string, error) {
	t.mu.Lock()
	t.called = true
	t.callArgs = args.ToJSON()
	cancel := t.cancel
	t.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	// Return empty: generation has just been cancelled, so FM won't see this result.
	return "", nil
}

// wasCalled returns whether this tool was called during generation.
func (t *proxyTool) wasCalled() (bool, string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.called, t.callArgs
}

// buildProxyTools converts OpenAI tool definitions into proxy fm.Tool instances.
func buildProxyTools(tools []ToolDefinition, cancel context.CancelFunc) ([]*proxyTool, error) {
	var proxyTools []*proxyTool

	for _, td := range tools {
		if td.Type != "function" {
			continue
		}

		schema, err := schemaFromJSONParams(td.Function.Name, td.Function.Parameters)
		if err != nil {
			return nil, fmt.Errorf("invalid parameters for tool %q: %w", td.Function.Name, err)
		}

		pt := &proxyTool{
			toolName: td.Function.Name,
			toolDesc: td.Function.Description,
			schema:   schema,
			cancel:   cancel,
		}
		proxyTools = append(proxyTools, pt)
	}

	return proxyTools, nil
}

// proxyToolsToFMTools converts proxy tools to the fm.Tool interface slice.
func proxyToolsToFMTools(pts []*proxyTool) []fm.Tool {
	tools := make([]fm.Tool, len(pts))
	for i, pt := range pts {
		tools[i] = pt
	}
	return tools
}

// schemaFromJSONParams converts an OpenAI JSON Schema parameters object
// into an fm.GenerationSchema.
func schemaFromJSONParams(name string, raw json.RawMessage) (*fm.GenerationSchema, error) {
	if len(raw) == 0 {
		return fm.NewGenerationSchema(name, ""), nil
	}

	var params map[string]any
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters JSON: %w", err)
	}

	schema := fm.NewGenerationSchema(name, "")

	properties, _ := params["properties"].(map[string]any)
	required := toStringSet(params["required"])

	for propName, propVal := range properties {
		propDef, ok := propVal.(map[string]any)
		if !ok {
			continue
		}

		typeName := mapJSONSchemaType(propDef)
		desc, _ := propDef["description"].(string)
		isOptional := !required[propName]

		prop := fm.NewProperty(propName, desc, typeName, isOptional)

		if enumVals, ok := propDef["enum"].([]any); ok {
			choices := make([]string, 0, len(enumVals))
			for _, v := range enumVals {
				choices = append(choices, fmt.Sprintf("%v", v))
			}
			if len(choices) > 0 {
				prop.AddAnyOfGuide(choices, false)
			}
		}

		if minV, ok := propDef["minimum"].(float64); ok {
			prop.AddMinimumGuide(minV, false)
		}
		if maxV, ok := propDef["maximum"].(float64); ok {
			prop.AddMaximumGuide(maxV, false)
		}

		schema.AddProperty(prop)
	}

	return schema, nil
}

// mapJSONSchemaType converts a JSON Schema type to an FM type name.
func mapJSONSchemaType(propDef map[string]any) string {
	t, _ := propDef["type"].(string)
	switch t {
	case "string":
		return "string"
	case "integer":
		return "integer"
	case "number":
		return "float"
	case "boolean":
		return "boolean"
	case "array":
		if items, ok := propDef["items"].(map[string]any); ok {
			return "array<" + mapJSONSchemaType(items) + ">"
		}
		return "array<string>"
	default:
		return "string"
	}
}

// toStringSet converts an any (expected []any of strings) to a set.
func toStringSet(v any) map[string]bool {
	set := make(map[string]bool)
	arr, ok := v.([]any)
	if !ok {
		return set
	}
	for _, item := range arr {
		if s, ok := item.(string); ok {
			set[s] = true
		}
	}
	return set
}

// hasToolResults checks if any message has the "tool" role.
func hasToolResults(messages []Message) bool {
	for _, m := range messages {
		if m.Role == "tool" {
			return true
		}
	}
	return false
}
