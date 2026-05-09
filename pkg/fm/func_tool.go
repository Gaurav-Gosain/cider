package fm

import (
	"encoding/json"
	"fmt"
)

// FuncTool creates a Tool from a function that accepts a typed struct.
// The schema is automatically generated from T's struct tags.
//
// Example:
//
//	type WeatherArgs struct {
//	    Location string `json:"location" description:"City name"`
//	    Unit     string `json:"unit"     description:"Temperature unit" enum:"celsius,fahrenheit"`
//	}
//
//	weatherTool := fm.FuncTool("get_weather", "Get current weather", func(args WeatherArgs) (string, error) {
//	    return fmt.Sprintf("72°F in %s", args.Location), nil
//	})
func FuncTool[T any](name, description string, fn func(T) (string, error)) Tool {
	schema := SchemaFor[T]()
	return &funcTool[T]{
		name:   name,
		desc:   description,
		schema: schema,
		fn:     fn,
	}
}

type funcTool[T any] struct {
	name   string
	desc   string
	schema *GenerationSchema
	fn     func(T) (string, error)
}

func (t *funcTool[T]) Name() string                      { return t.name }
func (t *funcTool[T]) Description() string               { return t.desc }
func (t *funcTool[T]) ArgumentsSchema() *GenerationSchema { return t.schema }

func (t *funcTool[T]) Call(args *GeneratedContent) (string, error) {
	jsonStr := args.ToJSON()
	var parsed T
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}
	return t.fn(parsed)
}
