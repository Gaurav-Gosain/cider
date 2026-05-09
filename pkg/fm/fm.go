package fm

import (
	"sync"
	"sync/atomic"
)

// callbackRegistry maps unique IDs to callback channels for bridging C callbacks to Go.
var (
	callbackRegistry sync.Map
	nextCallbackID   atomic.Uint64
)

type callbackInfo struct {
	resultCh chan callbackResult
}

type callbackResult struct {
	status  int
	content string
	err     error
}

type structuredCallbackInfo struct {
	resultCh chan structuredCallbackResult
}

type structuredCallbackResult struct {
	status  int
	content uintptr
	err     error
}

type streamCallbackInfo struct {
	chunkCh chan string
	doneCh  chan error
}

func registerCallback(info any) uint64 {
	id := nextCallbackID.Add(1)
	callbackRegistry.Store(id, info)
	return id
}

func unregisterCallback(id uint64) {
	callbackRegistry.Delete(id)
}

func lookupCallback(id uint64) (any, bool) {
	return callbackRegistry.Load(id)
}
