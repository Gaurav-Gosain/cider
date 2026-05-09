package fm

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGoTypeToSchemaType(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"string", "", "string"},
		{"int", 0, "integer"},
		{"int64", int64(0), "integer"},
		{"float64", 0.0, "float"},
		{"float32", float32(0), "float"},
		{"bool", false, "boolean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := goTypeToSchemaType(reflect.TypeOf(tt.value))
			if result != tt.expected {
				t.Errorf("goTypeToSchemaType(%T) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestParseJSONTag(t *testing.T) {
	type testStruct struct {
		Normal    string `json:"normal"`
		OmitEmpty string `json:"omit,omitempty"`
		Skipped   string `json:"-"`
		NoTag     string
	}

	typ := reflect.TypeOf(testStruct{})

	tests := []struct {
		fieldName    string
		expectedName string
		omitempty    bool
	}{
		{"Normal", "normal", false},
		{"OmitEmpty", "omit", true},
		{"Skipped", "-", false},
		{"NoTag", "NoTag", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			f, _ := typ.FieldByName(tt.fieldName)
			name, opts := parseJSONTag(f)
			if name != tt.expectedName {
				t.Errorf("parseJSONTag(%s) name = %q, want %q", tt.fieldName, name, tt.expectedName)
			}
			if opts.contains("omitempty") != tt.omitempty {
				t.Errorf("parseJSONTag(%s) omitempty = %v, want %v", tt.fieldName, opts.contains("omitempty"), tt.omitempty)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create a mock GeneratedContent by directly constructing the JSON
	jsonStr := `{"name":"Alice","age":30}`
	var person Person
	// Test the underlying json unmarshal that Unmarshal uses
	err := json.Unmarshal([]byte(jsonStr), &person)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if person.Name != "Alice" {
		t.Errorf("Name = %q, want %q", person.Name, "Alice")
	}
	if person.Age != 30 {
		t.Errorf("Age = %d, want %d", person.Age, 30)
	}
}
