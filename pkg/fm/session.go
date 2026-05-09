package fm

import (
	"context"
	"runtime"
	"sync"
)

// Session represents a conversation session with the Foundation Model.
type Session struct {
	ptr          uintptr
	mu           sync.Mutex
	tools        []Tool
	bridgedTools []*BridgedTool
}

// GenerationOptions controls response generation behavior.
type GenerationOptions struct {
	// Temperature controls randomness (0.0-2.0). Negative means use default.
	Temperature float64
	// MaxTokens limits the maximum response tokens. 0 means use default.
	MaxTokens int
}

// SessionOption configures a Session.
type SessionOption func(*sessionConfig)

type sessionConfig struct {
	instructions *string
	model        *SystemLanguageModel
	tools        []Tool
}

// WithInstructions sets the system instructions for the session.
func WithInstructions(instructions string) SessionOption {
	return func(c *sessionConfig) {
		c.instructions = &instructions
	}
}

// WithModel sets the SystemLanguageModel for the session.
func WithModel(model *SystemLanguageModel) SessionOption {
	return func(c *sessionConfig) {
		c.model = model
	}
}

// WithTools sets the tools available to the model during the session.
func WithTools(tools ...Tool) SessionOption {
	return func(c *sessionConfig) {
		c.tools = tools
	}
}

// NewSession creates a new conversation session with the Foundation Model.
func NewSession(opts ...SessionOption) (*Session, error) {
	cfg := &sessionConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.model == nil && cfg.instructions == nil && len(cfg.tools) == 0 {
		ptr := fmLanguageModelSessionCreateDefault()
		s := &Session{ptr: ptr}
		runtime.SetFinalizer(s, func(s *Session) { s.Close() })
		return s, nil
	}

	var modelRef uintptr
	if cfg.model != nil {
		modelRef = cfg.model.ptr
	}

	var instrPtr uintptr
	var instrBuf []byte
	if cfg.instructions != nil {
		instrPtr, instrBuf = cStringPtr(*cfg.instructions)
	}

	toolCount := len(cfg.tools)
	var bridgedTools []*BridgedTool
	var cToolsArg uintptr
	var cToolSlice []uintptr

	if toolCount > 0 {
		bridgedTools = make([]*BridgedTool, toolCount)
		cToolSlice = make([]uintptr, toolCount)
		for i, t := range cfg.tools {
			bt, err := newBridgedTool(t)
			if err != nil {
				for j := 0; j < i; j++ {
					bridgedTools[j].Close()
				}
				return nil, err
			}
			bridgedTools[i] = bt
			cToolSlice[i] = bt.ptr
		}
		cToolsArg = uintptr(ptrToSliceData(cToolSlice))
	}

	ptr := fmLanguageModelSessionCreateFromSystemLanguageModel(
		modelRef, instrPtr, cToolsArg, toolCount,
	)

	_ = instrBuf   // keep alive
	_ = cToolSlice // keep alive

	s := &Session{
		ptr:          ptr,
		tools:        cfg.tools,
		bridgedTools: bridgedTools,
	}
	runtime.SetFinalizer(s, func(s *Session) { s.Close() })
	return s, nil
}

// Respond sends a prompt and returns the complete response.
// Optional GenerationOptions can be passed to control temperature and max tokens.
func (s *Session) Respond(ctx context.Context, prompt string, opts ...GenerationOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	promptPtr, promptBuf := cStringPtr(prompt)

	info := &callbackInfo{
		resultCh: make(chan callbackResult, 1),
	}
	id := registerCallback(info)
	defer unregisterCallback(id)

	var taskRef uintptr
	if len(opts) > 0 && (opts[0].Temperature >= 0 || opts[0].MaxTokens > 0) {
		o := opts[0]
		temp := o.Temperature
		if temp < 0 {
			temp = -1 // signal "use default" to Swift
		}
		taskRef = fmLanguageModelSessionRespondWithOptions(
			s.ptr, promptPtr, temp, o.MaxTokens, uintptr(id), responseCallbackPtr,
		)
	} else {
		taskRef = fmLanguageModelSessionRespond(
			s.ptr, promptPtr, uintptr(id), responseCallbackPtr,
		)
	}
	_ = promptBuf // keep alive

	select {
	case <-ctx.Done():
		fmTaskCancel(taskRef)
		return "", ctx.Err()
	case result := <-info.resultCh:
		if result.err != nil {
			return "", result.err
		}
		return result.content, nil
	}
}

// StreamResponse sends a prompt and returns channels for streaming chunks and completion.
// Optional GenerationOptions can be passed to control temperature and max tokens.
func (s *Session) StreamResponse(ctx context.Context, prompt string, opts ...GenerationOptions) (<-chan string, <-chan error) {
	chunkCh := make(chan string, 64)
	errCh := make(chan error, 1)

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		defer close(chunkCh)
		defer close(errCh)

		promptPtr, promptBuf := cStringPtr(prompt)

		var streamRef uintptr
		if len(opts) > 0 && (opts[0].Temperature >= 0 || opts[0].MaxTokens > 0) {
			o := opts[0]
			temp := o.Temperature
			if temp < 0 {
				temp = -1
			}
			streamRef = fmLanguageModelSessionStreamResponseWithOptions(s.ptr, promptPtr, temp, o.MaxTokens)
		} else {
			streamRef = fmLanguageModelSessionStreamResponse(s.ptr, promptPtr)
		}
		_ = promptBuf // keep alive

		info := &streamCallbackInfo{
			chunkCh: chunkCh,
			doneCh:  make(chan error, 1),
		}
		id := registerCallback(info)
		defer unregisterCallback(id)

		go fmLanguageModelSessionResponseStreamIterate(
			streamRef, uintptr(id), streamCallbackPtr,
		)

		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
		case err := <-info.doneCh:
			if err != nil {
				errCh <- err
			}
		}
	}()

	return chunkCh, errCh
}

// RespondWithSchema sends a prompt and returns structured content based on the schema.
func (s *Session) RespondWithSchema(ctx context.Context, prompt string, schema *GenerationSchema) (*GeneratedContent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	promptPtr, promptBuf := cStringPtr(prompt)

	info := &structuredCallbackInfo{
		resultCh: make(chan structuredCallbackResult, 1),
	}
	id := registerCallback(info)
	defer unregisterCallback(id)

	taskRef := fmLanguageModelSessionRespondWithSchema(
		s.ptr, promptPtr, schema.ptr, uintptr(id), structuredResponseCallbackPtr,
	)
	_ = promptBuf

	select {
	case <-ctx.Done():
		fmTaskCancel(taskRef)
		return nil, ctx.Err()
	case result := <-info.resultCh:
		if result.err != nil {
			return nil, result.err
		}
		return newGeneratedContent(result.content), nil
	}
}

// RespondWithJSONSchema sends a prompt and returns structured content based on a JSON schema string.
func (s *Session) RespondWithJSONSchema(ctx context.Context, prompt string, jsonSchema string) (*GeneratedContent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	promptPtr, promptBuf := cStringPtr(prompt)
	schemaPtr, schemaBuf := cStringPtr(jsonSchema)

	info := &structuredCallbackInfo{
		resultCh: make(chan structuredCallbackResult, 1),
	}
	id := registerCallback(info)
	defer unregisterCallback(id)

	taskRef := fmLanguageModelSessionRespondWithSchemaFromJSON(
		s.ptr, promptPtr, schemaPtr, uintptr(id), structuredResponseCallbackPtr,
	)
	_ = promptBuf
	_ = schemaBuf

	select {
	case <-ctx.Done():
		fmTaskCancel(taskRef)
		return nil, ctx.Err()
	case result := <-info.resultCh:
		if result.err != nil {
			return nil, result.err
		}
		return newGeneratedContent(result.content), nil
	}
}

// IsResponding returns whether the session is currently generating a response.
func (s *Session) IsResponding() bool {
	return fmLanguageModelSessionIsResponding(s.ptr)
}

// Reset resets the session's task state.
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmLanguageModelSessionReset(s.ptr)
}

// Close releases the underlying C reference and cleans up bridged tools.
func (s *Session) Close() {
	// Close bridged tools first to remove them from the global registry
	for _, bt := range s.bridgedTools {
		bt.Close()
	}
	s.bridgedTools = nil

	if s.ptr != 0 {
		release(s.ptr)
		s.ptr = 0
	}
}
