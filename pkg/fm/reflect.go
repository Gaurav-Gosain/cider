package fm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// SchemaFor generates a GenerationSchema from a Go struct type.
// Struct fields are mapped using the `json` tag for the property name
// and the `description` tag for the property description.
//
// Supported field types: string, int/int64 ("integer"), float64/float32 ("float"),
// bool ("boolean").
//
// Fields with `json:"-"` are skipped. Fields with `json:",omitempty"` are marked optional.
//
// Example:
//
//	type Person struct {
//	    Name       string  `json:"name"        description:"The person's full name"`
//	    Age        int     `json:"age"         description:"The person's age"`
//	    Occupation string  `json:"occupation"  description:"The person's job title"`
//	}
//	schema := fm.SchemaFor[Person]()
func SchemaFor[T any]() *GenerationSchema {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("fm.SchemaFor: expected struct, got %s", t.Kind()))
	}

	schema := NewGenerationSchema(t.Name(), "")

	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name, opts := parseJSONTag(f)
		if name == "-" {
			continue
		}

		desc := f.Tag.Get("description")
		typeName := goTypeToSchemaType(f.Type)
		isOptional := opts.contains("omitempty")

		prop := NewProperty(name, desc, typeName, isOptional)

		// Handle enum tag
		if enumTag := f.Tag.Get("enum"); enumTag != "" {
			choices := strings.Split(enumTag, ",")
			prop.AddAnyOfGuide(choices, false)
		}

		schema.AddProperty(prop)
	}

	return schema
}

// Unmarshal decodes a GeneratedContent into a Go value (typically a struct pointer).
// It uses JSON deserialization under the hood.
//
// Example:
//
//	var person Person
//	err := fm.Unmarshal(content, &person)
func Unmarshal(gc *GeneratedContent, dest any) error {
	jsonStr := gc.ToJSON()
	return json.Unmarshal([]byte(jsonStr), dest)
}

func goTypeToSchemaType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "boolean"
	default:
		return "string"
	}
}

type jsonTagOpts string

func (o jsonTagOpts) contains(optName string) bool {
	for o != "" {
		var name string
		if idx := strings.Index(string(o), ","); idx >= 0 {
			name, o = string(o)[:idx], o[idx+1:]
		} else {
			name, o = string(o), ""
		}
		if name == optName {
			return true
		}
	}
	return false
}

func parseJSONTag(f reflect.StructField) (string, jsonTagOpts) {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name, ""
	}
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], jsonTagOpts(tag[idx+1:])
	}
	return tag, ""
}
