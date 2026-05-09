package fm

import (
	"fmt"
	"runtime"

	"github.com/ebitengine/purego"
)

// Tool is the interface for tools that can be used by the Foundation Model.
type Tool interface {
	Name() string
	Description() string
	ArgumentsSchema() *GenerationSchema
	Call(args *GeneratedContent) (string, error)
}

// BridgedTool wraps a Tool for use with the C API.
type BridgedTool struct {
	ptr         uintptr
	tool        Tool
	callbackPtr uintptr // unique per-tool C callback
}

func newBridgedTool(t Tool) (*BridgedTool, error) {
	namePtr, nameBuf := cStringPtr(t.Name())
	descPtr, descBuf := cStringPtr(t.Description())

	schema := t.ArgumentsSchema()

	// bt is allocated here and captured by the callback closure below.
	// This ensures each tool's callback dispatches to the correct Go tool.
	bt := &BridgedTool{tool: t}

	// Create a unique callback for this specific tool.
	// The Swift side stores this callback per BridgedTool and calls it
	// only when this specific tool is invoked by the model.
	cbPtr := purego.NewCallback(func(contentRef uintptr, callId uint32) {
		content := newGeneratedContent(contentRef)

		result, err := bt.tool.Call(content)
		if err != nil {
			result = "Error: " + err.Error()
		}

		resultPtr, resultBuf := cStringPtr(result)
		fmBridgedToolFinishCall(bt.ptr, callId, resultPtr)
		_ = resultBuf // keep alive
	})

	var errorCode int
	var errorDesc uintptr

	ptr := fmBridgedToolCreate(
		namePtr, descPtr, schema.ptr, cbPtr, &errorCode, &errorDesc,
	)
	_ = nameBuf
	_ = descBuf

	if err := extractError(errorCode, errorDesc); err != nil {
		return nil, fmt.Errorf("failed to create bridged tool %q: %w", t.Name(), err)
	}

	bt.ptr = ptr
	bt.callbackPtr = cbPtr

	runtime.SetFinalizer(bt, func(bt *BridgedTool) { bt.Close() })
	return bt, nil
}

// Close releases the underlying C reference.
func (bt *BridgedTool) Close() {
	if bt.ptr != 0 {
		release(bt.ptr)
		bt.ptr = 0
	}
}
