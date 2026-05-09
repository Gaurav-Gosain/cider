package fm

import (
	"github.com/ebitengine/purego"
)

// Callback function pointers registered with purego.NewCallback.
// These bridge C callbacks into Go channels via the callback registry.
var (
	responseCallbackPtr           uintptr
	streamCallbackPtr             uintptr
	structuredResponseCallbackPtr uintptr
)

func init() {
	responseCallbackPtr = purego.NewCallback(responseCallbackFunc)
	streamCallbackPtr = purego.NewCallback(streamCallbackFunc)
	structuredResponseCallbackPtr = purego.NewCallback(structuredResponseCallbackFunc)
}

// responseCallbackFunc is called by the FM runtime when a text response is ready.
// Signature: void(int status, const char *content, size_t length, void *userInfo)
func responseCallbackFunc(status int, content uintptr, length uintptr, userInfo uintptr) {
	id := uint64(userInfo)
	val, ok := lookupCallback(id)
	if !ok {
		return
	}

	info, ok := val.(*callbackInfo)
	if !ok {
		return
	}

	text := goStringN(content, int(length))

	if status != 0 {
		info.resultCh <- callbackResult{status: status, err: statusCodeToError(status, text)}
	} else {
		info.resultCh <- callbackResult{status: status, content: text}
	}
}

// streamCallbackFunc is called by the FM runtime for each streaming chunk.
func streamCallbackFunc(status int, content uintptr, length uintptr, userInfo uintptr) {
	id := uint64(userInfo)
	val, ok := lookupCallback(id)
	if !ok {
		return
	}

	info, ok := val.(*streamCallbackInfo)
	if !ok {
		return
	}

	if status != 0 {
		text := goStringN(content, int(length))
		info.doneCh <- statusCodeToError(status, text)
		return
	}

	if content == 0 {
		info.doneCh <- nil
		return
	}

	text := goStringN(content, int(length))
	info.chunkCh <- text
}

// structuredResponseCallbackFunc is called by the FM runtime for structured responses.
// Signature: void(int status, FMGeneratedContentRef content, void *userInfo)
func structuredResponseCallbackFunc(status int, content uintptr, userInfo uintptr) {
	id := uint64(userInfo)
	val, ok := lookupCallback(id)
	if !ok {
		return
	}

	info, ok := val.(*structuredCallbackInfo)
	if !ok {
		return
	}

	if status != 0 {
		info.resultCh <- structuredCallbackResult{status: status, err: errorFromCode(status, "")}
	} else {
		if content != 0 {
			retain(content)
		}
		info.resultCh <- structuredCallbackResult{status: status, content: content}
	}
}
