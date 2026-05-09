package fm

import (
	"encoding/json"
	"runtime"
)

// GeneratedContent wraps structured content returned by the Foundation Model.
type GeneratedContent struct {
	ptr uintptr
}

func newGeneratedContent(ptr uintptr) *GeneratedContent {
	if ptr == 0 {
		return nil
	}
	gc := &GeneratedContent{ptr: ptr}
	runtime.SetFinalizer(gc, func(gc *GeneratedContent) { gc.Close() })
	return gc
}

// ContentFromJSON creates a GeneratedContent from a JSON string.
func ContentFromJSON(jsonStr string) (*GeneratedContent, error) {
	cs, buf := cStringPtr(jsonStr)

	var errorCode int
	var errorDesc uintptr

	ptr := fmGeneratedContentCreateFromJSON(cs, &errorCode, &errorDesc)
	_ = buf

	if err := extractError(errorCode, errorDesc); err != nil {
		return nil, err
	}

	return newGeneratedContent(ptr), nil
}

// ToJSON returns the content as a JSON string.
func (gc *GeneratedContent) ToJSON() string {
	cs := fmGeneratedContentGetJSONString(gc.ptr)
	if cs == 0 {
		return "{}"
	}
	defer freeStr(cs)
	return goString(cs)
}

// ToMap returns the content as a map[string]any.
func (gc *GeneratedContent) ToMap() (map[string]any, error) {
	jsonStr := gc.ToJSON()
	var result map[string]any
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// PropertyValue returns the value of a named property as a string.
func (gc *GeneratedContent) PropertyValue(name string) (string, error) {
	namePtr, nameBuf := cStringPtr(name)

	var errorCode int
	var errorDesc uintptr

	cs := fmGeneratedContentGetPropertyValue(gc.ptr, namePtr, &errorCode, &errorDesc)
	_ = nameBuf

	if err := extractError(errorCode, errorDesc); err != nil {
		return "", err
	}
	if cs == 0 {
		return "", nil
	}
	defer freeStr(cs)
	return goString(cs), nil
}

// IsComplete returns whether the generation completed successfully.
func (gc *GeneratedContent) IsComplete() bool {
	return fmGeneratedContentIsComplete(gc.ptr)
}

// Close releases the underlying C reference.
func (gc *GeneratedContent) Close() {
	if gc.ptr != 0 {
		release(gc.ptr)
		gc.ptr = 0
	}
}
