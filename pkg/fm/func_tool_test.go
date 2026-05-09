package fm

import (
	"testing"
)

func TestFuncToolInterface(t *testing.T) {
	requireLib(t)

	type Args struct {
		City string `json:"city" description:"City name"`
		Unit string `json:"unit" description:"Temperature unit" enum:"celsius,fahrenheit"`
	}

	tool := FuncTool("weather", "Get weather info", func(args Args) (string, error) {
		return "sunny in " + args.City, nil
	})

	if tool.Name() != "weather" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "weather")
	}
	if tool.Description() != "Get weather info" {
		t.Errorf("Description() = %q, want %q", tool.Description(), "Get weather info")
	}
	if tool.ArgumentsSchema() == nil {
		t.Fatal("ArgumentsSchema() returned nil")
	}
}

func TestSchemaForPanicsOnNonStruct(t *testing.T) {
	// This test doesn't need the library since it panics before calling C
	defer func() {
		if r := recover(); r == nil {
			t.Error("SchemaFor[string]() did not panic")
		}
	}()
	SchemaFor[string]()
}

func TestSchemaForBasicStruct(t *testing.T) {
	requireLib(t)

	type Person struct {
		Name string `json:"name" description:"Full name"`
		Age  int    `json:"age"  description:"Age in years"`
	}

	schema := SchemaFor[Person]()
	if schema == nil {
		t.Fatal("SchemaFor[Person]() returned nil")
	}
	if schema.ptr == 0 {
		t.Fatal("schema has nil ptr")
	}
}

func TestSchemaForOptionalField(t *testing.T) {
	requireLib(t)

	type Config struct {
		Required string `json:"required" description:"Required field"`
		Optional string `json:"optional,omitempty" description:"Optional field"`
	}

	schema := SchemaFor[Config]()
	if schema == nil {
		t.Fatal("SchemaFor[Config]() returned nil")
	}
}

func TestSchemaForSkippedField(t *testing.T) {
	requireLib(t)

	type Data struct {
		Visible string `json:"visible"`
		Hidden  string `json:"-"`
	}

	schema := SchemaFor[Data]()
	if schema == nil {
		t.Fatal("SchemaFor[Data]() returned nil")
	}
}

func TestSchemaForEnumTag(t *testing.T) {
	requireLib(t)

	type Choice struct {
		Color string `json:"color" description:"Pick a color" enum:"red,green,blue"`
	}

	schema := SchemaFor[Choice]()
	if schema == nil {
		t.Fatal("SchemaFor[Choice]() returned nil")
	}
}
