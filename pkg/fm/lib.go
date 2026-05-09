package fm

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// C function signatures registered via purego
var (
	// SystemLanguageModel
	fmSystemLanguageModelGetDefault func() uintptr
	fmSystemLanguageModelCreate     func(useCase int, guardrails int) uintptr
	fmSystemLanguageModelIsAvailable func(ref uintptr, reason *int) bool

	// Session
	fmLanguageModelSessionCreateDefault                func() uintptr
	fmLanguageModelSessionCreateFromSystemLanguageModel func(model uintptr, instructions uintptr, tools uintptr, toolCount int) uintptr
	fmLanguageModelSessionIsResponding                  func(session uintptr) bool
	fmLanguageModelSessionReset                         func(session uintptr)
	fmLanguageModelSessionRespond                       func(session uintptr, prompt uintptr, userInfo uintptr, callback uintptr) uintptr
	fmLanguageModelSessionRespondWithOptions            func(session uintptr, prompt uintptr, temperature float64, maxTokens int, userInfo uintptr, callback uintptr) uintptr
	fmLanguageModelSessionStreamResponse                func(session uintptr, prompt uintptr) uintptr
	fmLanguageModelSessionStreamResponseWithOptions     func(session uintptr, prompt uintptr, temperature float64, maxTokens int) uintptr
	fmLanguageModelSessionResponseStreamIterate         func(stream uintptr, userInfo uintptr, callback uintptr)

	// Transcript
	fmLanguageModelSessionGetTranscriptJSONString func(session uintptr, errorCode *int, errorDesc *uintptr) uintptr

	// Schema
	fmGenerationSchemaCreate              func(name uintptr, description uintptr) uintptr
	fmGenerationSchemaPropertyCreate      func(name uintptr, description uintptr, typeName uintptr, isOptional bool) uintptr
	fmGenerationSchemaPropertyAddAnyOfGuide func(property uintptr, anyOf uintptr, choiceCount int, wrapped bool)
	fmGenerationSchemaPropertyAddCountGuide func(property uintptr, count int, wrapped bool)
	fmGenerationSchemaPropertyAddMaximumGuide func(property uintptr, maximum float64, wrapped bool)
	fmGenerationSchemaPropertyAddMinimumGuide func(property uintptr, minimum float64, wrapped bool)
	fmGenerationSchemaPropertyAddMinItemsGuide func(property uintptr, minItems int)
	fmGenerationSchemaPropertyAddMaxItemsGuide func(property uintptr, maxItems int)
	fmGenerationSchemaPropertyAddRangeGuide    func(property uintptr, minValue float64, maxValue float64, wrapped bool)
	fmGenerationSchemaPropertyAddRegex         func(property uintptr, pattern uintptr, wrapped bool)
	fmGenerationSchemaAddProperty              func(schema uintptr, property uintptr)
	fmGenerationSchemaAddReferenceSchema       func(schema uintptr, referenceSchema uintptr)
	fmGenerationSchemaGetJSONString            func(schema uintptr, errorCode *int, errorDesc *uintptr) uintptr

	// GeneratedContent
	fmGeneratedContentCreateFromJSON    func(jsonString uintptr, errorCode *int, errorDesc *uintptr) uintptr
	fmGeneratedContentGetJSONString     func(content uintptr) uintptr
	fmGeneratedContentGetPropertyValue  func(content uintptr, propertyName uintptr, errorCode *int, errorDesc *uintptr) uintptr
	fmGeneratedContentIsComplete        func(content uintptr) bool

	// Structured responses
	fmLanguageModelSessionRespondWithSchema         func(session uintptr, prompt uintptr, schema uintptr, userInfo uintptr, callback uintptr) uintptr
	fmLanguageModelSessionRespondWithSchemaFromJSON func(session uintptr, prompt uintptr, schemaJSON uintptr, userInfo uintptr, callback uintptr) uintptr

	// Tools
	fmBridgedToolCreate     func(name uintptr, description uintptr, parameters uintptr, callable uintptr, errorCode *int, errorDesc *uintptr) uintptr
	fmBridgedToolFinishCall func(tool uintptr, callId uint32, output uintptr)

	// Memory / task
	fmTaskCancel func(task uintptr)
	fmRetain     func(object uintptr)
	fmRelease    func(object uintptr)
	fmFreeString func(str uintptr)
)

var libHandle uintptr

// Init loads the Foundation Models dynamic library.
// It searches for libFoundationModels.dylib in (in order):
//   1. caller-supplied paths
//   2. $CIDER_LIB_PATH
//   3. the directory containing the executable, with symlinks resolved
//      (so brew installs that wrap the binary in libexec/ still find the
//      sibling dylib)
//   4. ../lib/libFoundationModels.dylib relative to the executable
//      (homebrew layout: bin/cider next to lib/libFoundationModels.dylib)
//   5. ./foundation-models-c/.build/arm64-apple-macosx/release/ (dev layout)
//   6. the bare library name, so dlopen falls back to DYLD_LIBRARY_PATH /
//      the system paths.
func Init(paths ...string) error {
	const dylib = "libFoundationModels.dylib"
	searchPaths := append([]string{}, paths...)

	if envPath := os.Getenv("CIDER_LIB_PATH"); envPath != "" {
		searchPaths = append(searchPaths, filepath.Join(envPath, dylib))
	}

	if exe, err := os.Executable(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(filepath.Dir(exe), dylib))
		if resolved, err := filepath.EvalSymlinks(exe); err == nil && resolved != exe {
			dir := filepath.Dir(resolved)
			searchPaths = append(searchPaths,
				filepath.Join(dir, dylib),
				filepath.Join(dir, "..", "lib", dylib),
			)
		}
	}

	searchPaths = append(searchPaths,
		"foundation-models-c/.build/arm64-apple-macosx/release/"+dylib,
		dylib,
	)

	var lastErr error
	for _, p := range searchPaths {
		handle, err := purego.Dlopen(p, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			libHandle = handle
			registerFunctions(handle)
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to load %s: %w", dylib, lastErr)
}

func registerFunctions(lib uintptr) {
	// SystemLanguageModel
	purego.RegisterLibFunc(&fmSystemLanguageModelGetDefault, lib, "FMSystemLanguageModelGetDefault")
	purego.RegisterLibFunc(&fmSystemLanguageModelCreate, lib, "FMSystemLanguageModelCreate")
	purego.RegisterLibFunc(&fmSystemLanguageModelIsAvailable, lib, "FMSystemLanguageModelIsAvailable")

	// Session
	purego.RegisterLibFunc(&fmLanguageModelSessionCreateDefault, lib, "FMLanguageModelSessionCreateDefault")
	purego.RegisterLibFunc(&fmLanguageModelSessionCreateFromSystemLanguageModel, lib, "FMLanguageModelSessionCreateFromSystemLanguageModel")
	purego.RegisterLibFunc(&fmLanguageModelSessionIsResponding, lib, "FMLanguageModelSessionIsResponding")
	purego.RegisterLibFunc(&fmLanguageModelSessionReset, lib, "FMLanguageModelSessionReset")
	purego.RegisterLibFunc(&fmLanguageModelSessionRespond, lib, "FMLanguageModelSessionRespond")
	purego.RegisterLibFunc(&fmLanguageModelSessionRespondWithOptions, lib, "FMLanguageModelSessionRespondWithOptions")
	purego.RegisterLibFunc(&fmLanguageModelSessionStreamResponse, lib, "FMLanguageModelSessionStreamResponse")
	purego.RegisterLibFunc(&fmLanguageModelSessionStreamResponseWithOptions, lib, "FMLanguageModelSessionStreamResponseWithOptions")
	purego.RegisterLibFunc(&fmLanguageModelSessionResponseStreamIterate, lib, "FMLanguageModelSessionResponseStreamIterate")

	// Transcript
	purego.RegisterLibFunc(&fmLanguageModelSessionGetTranscriptJSONString, lib, "FMLanguageModelSessionGetTranscriptJSONString")

	// Schema
	purego.RegisterLibFunc(&fmGenerationSchemaCreate, lib, "FMGenerationSchemaCreate")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyCreate, lib, "FMGenerationSchemaPropertyCreate")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddAnyOfGuide, lib, "FMGenerationSchemaPropertyAddAnyOfGuide")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddCountGuide, lib, "FMGenerationSchemaPropertyAddCountGuide")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddMaximumGuide, lib, "FMGenerationSchemaPropertyAddMaximumGuide")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddMinimumGuide, lib, "FMGenerationSchemaPropertyAddMinimumGuide")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddMinItemsGuide, lib, "FMGenerationSchemaPropertyAddMinItemsGuide")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddMaxItemsGuide, lib, "FMGenerationSchemaPropertyAddMaxItemsGuide")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddRangeGuide, lib, "FMGenerationSchemaPropertyAddRangeGuide")
	purego.RegisterLibFunc(&fmGenerationSchemaPropertyAddRegex, lib, "FMGenerationSchemaPropertyAddRegex")
	purego.RegisterLibFunc(&fmGenerationSchemaAddProperty, lib, "FMGenerationSchemaAddProperty")
	purego.RegisterLibFunc(&fmGenerationSchemaAddReferenceSchema, lib, "FMGenerationSchemaAddReferenceSchema")
	purego.RegisterLibFunc(&fmGenerationSchemaGetJSONString, lib, "FMGenerationSchemaGetJSONString")

	// GeneratedContent
	purego.RegisterLibFunc(&fmGeneratedContentCreateFromJSON, lib, "FMGeneratedContentCreateFromJSON")
	purego.RegisterLibFunc(&fmGeneratedContentGetJSONString, lib, "FMGeneratedContentGetJSONString")
	purego.RegisterLibFunc(&fmGeneratedContentGetPropertyValue, lib, "FMGeneratedContentGetPropertyValue")
	purego.RegisterLibFunc(&fmGeneratedContentIsComplete, lib, "FMGeneratedContentIsComplete")

	// Structured responses
	purego.RegisterLibFunc(&fmLanguageModelSessionRespondWithSchema, lib, "FMLanguageModelSessionRespondWithSchema")
	purego.RegisterLibFunc(&fmLanguageModelSessionRespondWithSchemaFromJSON, lib, "FMLanguageModelSessionRespondWithSchemaFromJSON")

	// Tools
	purego.RegisterLibFunc(&fmBridgedToolCreate, lib, "FMBridgedToolCreate")
	purego.RegisterLibFunc(&fmBridgedToolFinishCall, lib, "FMBridgedToolFinishCall")

	// Memory / task
	purego.RegisterLibFunc(&fmTaskCancel, lib, "FMTaskCancel")
	purego.RegisterLibFunc(&fmRetain, lib, "FMRetain")
	purego.RegisterLibFunc(&fmRelease, lib, "FMRelease")
	purego.RegisterLibFunc(&fmFreeString, lib, "FMFreeString")
}

// cStringPtr creates a C string and returns its pointer as uintptr.
// The caller must keep the byte slice alive until the C function returns.
func cStringPtr(s string) (uintptr, []byte) {
	b := append([]byte(s), 0)
	return uintptr(unsafe.Pointer(&b[0])), b
}

// goString converts a C string pointer to a Go string.
func goString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	// Find null terminator
	var length int
	for {
		b := *(*byte)(unsafe.Pointer(ptr + uintptr(length)))
		if b == 0 {
			break
		}
		length++
	}
	if length == 0 {
		return ""
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), length))
}

// goStringN converts a C string pointer with known length to a Go string.
func goStringN(ptr uintptr, length int) string {
	if ptr == 0 || length == 0 {
		return ""
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), length))
}

// retain increments the reference count of a Foundation Models object.
func retain(ptr uintptr) {
	if ptr != 0 {
		fmRetain(ptr)
	}
}

// release decrements the reference count of a Foundation Models object.
func release(ptr uintptr) {
	if ptr != 0 {
		fmRelease(ptr)
	}
}

// freeStr frees a C string allocated by the Foundation Models library.
func freeStr(ptr uintptr) {
	if ptr != 0 {
		fmFreeString(ptr)
	}
}

// ptrToSliceData returns a pointer to the first element of a uintptr slice.
func ptrToSliceData(s []uintptr) unsafe.Pointer {
	return unsafe.Pointer(&s[0])
}

// Keep runtime imported for KeepAlive usage in other files
var _ = runtime.KeepAlive
