package fm

import (
	"runtime"
	"unsafe"
)

// GenerationSchema defines the structure for guided/structured generation.
type GenerationSchema struct {
	ptr uintptr
}

// NewGenerationSchema creates a new generation schema.
func NewGenerationSchema(name, description string) *GenerationSchema {
	namePtr, nameBuf := cStringPtr(name)
	var descPtr uintptr
	var descBuf []byte
	if description != "" {
		descPtr, descBuf = cStringPtr(description)
	}

	ptr := fmGenerationSchemaCreate(namePtr, descPtr)
	_ = nameBuf
	_ = descBuf

	gs := &GenerationSchema{ptr: ptr}
	runtime.SetFinalizer(gs, func(gs *GenerationSchema) { gs.Close() })
	return gs
}

// AddProperty adds a property to the schema.
func (gs *GenerationSchema) AddProperty(prop *Property) {
	fmGenerationSchemaAddProperty(gs.ptr, prop.ptr)
}

// AddReferenceSchema adds a nested reference schema.
func (gs *GenerationSchema) AddReferenceSchema(ref *GenerationSchema) {
	fmGenerationSchemaAddReferenceSchema(gs.ptr, ref.ptr)
}

// ToJSON returns the schema as a JSON string.
func (gs *GenerationSchema) ToJSON() (string, error) {
	var errorCode int
	var errorDesc uintptr

	cs := fmGenerationSchemaGetJSONString(gs.ptr, &errorCode, &errorDesc)
	if err := extractError(errorCode, errorDesc); err != nil {
		return "", err
	}
	if cs == 0 {
		return "", nil
	}
	defer freeStr(cs)
	return goString(cs), nil
}

// Close releases the underlying C reference.
func (gs *GenerationSchema) Close() {
	if gs.ptr != 0 {
		release(gs.ptr)
		gs.ptr = 0
	}
}

// Property represents a property in a generation schema.
type Property struct {
	ptr uintptr
}

// NewProperty creates a new schema property.
func NewProperty(name, description, typeName string, isOptional bool) *Property {
	namePtr, nameBuf := cStringPtr(name)
	var descPtr uintptr
	var descBuf []byte
	if description != "" {
		descPtr, descBuf = cStringPtr(description)
	}
	typePtr, typeBuf := cStringPtr(typeName)

	ptr := fmGenerationSchemaPropertyCreate(namePtr, descPtr, typePtr, isOptional)
	_ = nameBuf
	_ = descBuf
	_ = typeBuf

	p := &Property{ptr: ptr}
	runtime.SetFinalizer(p, func(p *Property) { p.Close() })
	return p
}

// AddAnyOfGuide constrains the property to one of the given values.
func (p *Property) AddAnyOfGuide(choices []string, wrapped bool) {
	// Build a C array of C strings
	cPtrs := make([]uintptr, len(choices))
	bufs := make([][]byte, len(choices))
	for i, c := range choices {
		cPtrs[i], bufs[i] = cStringPtr(c)
	}

	arrayPtr := uintptr(unsafe.Pointer(&cPtrs[0]))
	fmGenerationSchemaPropertyAddAnyOfGuide(p.ptr, arrayPtr, len(choices), wrapped)
	_ = bufs // keep alive
}

// AddCountGuide constrains the count of items.
func (p *Property) AddCountGuide(count int, wrapped bool) {
	fmGenerationSchemaPropertyAddCountGuide(p.ptr, count, wrapped)
}

// AddMaximumGuide constrains the maximum value.
func (p *Property) AddMaximumGuide(maximum float64, wrapped bool) {
	fmGenerationSchemaPropertyAddMaximumGuide(p.ptr, maximum, wrapped)
}

// AddMinimumGuide constrains the minimum value.
func (p *Property) AddMinimumGuide(minimum float64, wrapped bool) {
	fmGenerationSchemaPropertyAddMinimumGuide(p.ptr, minimum, wrapped)
}

// AddMinItemsGuide constrains the minimum number of items in an array.
func (p *Property) AddMinItemsGuide(minItems int) {
	fmGenerationSchemaPropertyAddMinItemsGuide(p.ptr, minItems)
}

// AddMaxItemsGuide constrains the maximum number of items in an array.
func (p *Property) AddMaxItemsGuide(maxItems int) {
	fmGenerationSchemaPropertyAddMaxItemsGuide(p.ptr, maxItems)
}

// AddRangeGuide constrains the value to a numeric range.
func (p *Property) AddRangeGuide(minValue, maxValue float64, wrapped bool) {
	fmGenerationSchemaPropertyAddRangeGuide(p.ptr, minValue, maxValue, wrapped)
}

// AddRegexGuide constrains the value to match a regex pattern.
func (p *Property) AddRegexGuide(pattern string, wrapped bool) {
	patternPtr, patternBuf := cStringPtr(pattern)
	fmGenerationSchemaPropertyAddRegex(p.ptr, patternPtr, wrapped)
	_ = patternBuf
}

// Close releases the underlying C reference.
func (p *Property) Close() {
	if p.ptr != 0 {
		release(p.ptr)
		p.ptr = 0
	}
}
