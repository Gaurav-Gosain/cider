package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"charm.land/log/v2"
	"github.com/Gaurav-Gosain/cider/pkg/fm"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const modelID = "apple-on-device-fm"

type handler struct {
	instructions string
}

func (h *handler) listModels(c *echo.Context) error {
	log.Debug("Listing models")
	resp := ModelsResponse{
		Object: "list",
		Data: []ModelObject{
			{
				ID:      modelID,
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "apple",
				ContextWindow: fm.ContextSize,
			},
		},
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) chatCompletions(c *echo.Context) error {
	var req ChatCompletionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: "Invalid request body: " + err.Error(),
				Type:    "invalid_request_error",
			},
		})
	}

	if len(req.Messages) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: "messages array is required and must not be empty",
				Type:    "invalid_request_error",
			},
		})
	}

	log.Debug("Chat completion request", "model", req.Model, "messages", len(req.Messages), "stream", req.Stream, "tools", len(req.Tools))

	// If tools are defined and no tool results have come back yet, run the
	// tool-interception path so we can hand tool calls back to the client.
	if len(req.Tools) > 0 && !hasToolResults(req.Messages) {
		return h.handleWithTools(c, req)
	}

	instructions, prompt := h.buildPrompt(req.Messages)

	var opts []fm.SessionOption
	if instructions != "" {
		opts = append(opts, fm.WithInstructions(instructions))
	} else if h.instructions != "" {
		opts = append(opts, fm.WithInstructions(h.instructions))
	}

	session, err := fm.NewSession(opts...)
	if err != nil {
		log.Error("Failed to create session", "err", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Message: "Failed to create model session: " + err.Error(),
				Type:    "server_error",
			},
		})
	}
	defer session.Close()

	genOpts := buildGenerationOptions(req)

	if req.Stream {
		return h.handleStreamingResponse(c, session, prompt, genOpts)
	}
	return h.handleResponse(c, session, prompt, genOpts)
}

// handleWithTools creates a session with proxy tools that intercept tool
// calls from the model and return them to the OpenAI client instead of
// letting the model execute them.
func (h *handler) handleWithTools(c *echo.Context, req ChatCompletionRequest) error {
	ctx := c.Request().Context()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	proxyTools, err := buildProxyTools(req.Tools, cancel)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: "Invalid tool definition: " + err.Error(),
				Type:    "invalid_request_error",
			},
		})
	}

	instructions, prompt := h.buildPrompt(req.Messages)

	var opts []fm.SessionOption
	if instructions != "" {
		opts = append(opts, fm.WithInstructions(instructions))
	} else if h.instructions != "" {
		opts = append(opts, fm.WithInstructions(h.instructions))
	}
	opts = append(opts, fm.WithTools(proxyToolsToFMTools(proxyTools)...))

	session, err := fm.NewSession(opts...)
	if err != nil {
		log.Error("Failed to create session with tools", "err", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Message: "Failed to create model session: " + err.Error(),
				Type:    "server_error",
			},
		})
	}
	defer session.Close()

	// Non-streaming Respond keeps the tool-detection path simple: when a
	// proxy tool fires it cancels the context and Respond returns.
	response, respondErr := session.Respond(ctx, prompt)

	var toolCalls []ToolCall
	for _, pt := range proxyTools {
		called, args := pt.wasCalled()
		if called {
			log.Debug("Tool called", "name", pt.toolName, "args", args)
			toolCalls = append(toolCalls, ToolCall{
				ID:   "call_" + uuid.New().String(),
				Type: "function",
				Function: ToolCallFunction{
					Name:      pt.toolName,
					Arguments: args,
				},
			})
		}
	}

	if len(toolCalls) > 0 {
		completionID := "chatcmpl-" + uuid.New().String()
		log.Debug("Returning tool calls", "id", completionID, "count", len(toolCalls))

		if req.Stream {
			return h.handleStreamingToolCalls(c, completionID, toolCalls)
		}

		return c.JSON(http.StatusOK, ChatCompletionResponse{
			ID:      completionID,
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   modelID,
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:      "assistant",
						ToolCalls: toolCalls,
					},
					FinishReason: "tool_calls",
				},
			},
		})
	}

	// No tool was called: return the normal response.
	if respondErr != nil {
		return h.handleError(c, respondErr)
	}

	completionID := "chatcmpl-" + uuid.New().String()
	log.Debug("Completion response (with tools, no call)", "id", completionID, "length", len(response))

	resp := ChatCompletionResponse{
		ID:      completionID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelID,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: StringPtr(response),
				},
				FinishReason: "stop",
			},
		},
	}

	return c.JSON(http.StatusOK, resp)
}

// handleStreamingToolCalls sends tool calls as SSE chunks in OpenAI streaming format.
func (h *handler) handleStreamingToolCalls(c *echo.Context, completionID string, toolCalls []ToolCall) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	created := time.Now().Unix()

	var deltas []ToolCallDelta
	for i, tc := range toolCalls {
		deltas = append(deltas, ToolCallDelta{
			Index: i,
			ID:    tc.ID,
			Type:  "function",
			Function: ToolCallFunction{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}

	chunk := ChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   modelID,
		Choices: []ChunkChoice{
			{
				Index: 0,
				Delta: ChunkDelta{
					Role:      "assistant",
					ToolCalls: deltas,
				},
			},
		},
	}
	if err := writeSSE(c, chunk); err != nil {
		return err
	}

	finishReason := "tool_calls"
	finalChunk := ChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   modelID,
		Choices: []ChunkChoice{
			{
				Index:        0,
				Delta:        ChunkDelta{},
				FinishReason: &finishReason,
			},
		},
	}
	if err := writeSSE(c, finalChunk); err != nil {
		return err
	}

	fmt.Fprint(c.Response(), "data: [DONE]\n\n")
	flush(c)
	return nil
}

func (h *handler) buildPrompt(messages []Message) (instructions string, prompt string) {
	var systemParts []string
	var conversationParts []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemParts = append(systemParts, msg.ContentString())
		case "user":
			conversationParts = append(conversationParts, msg.ContentString())
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					conversationParts = append(conversationParts,
						fmt.Sprintf("[You previously called the tool %q with arguments: %s]",
							tc.Function.Name, tc.Function.Arguments))
				}
			} else {
				conversationParts = append(conversationParts, "[Assistant]: "+msg.ContentString())
			}
		case "tool":
			conversationParts = append(conversationParts,
				fmt.Sprintf("[Tool result]: %s", msg.ContentString()))
		}
	}

	if len(systemParts) > 0 {
		instructions = strings.Join(systemParts, "\n")
	}

	if len(conversationParts) > 0 {
		prompt = strings.Join(conversationParts, "\n")
	}

	return instructions, prompt
}

func (h *handler) handleResponse(c *echo.Context, session *fm.Session, prompt string, opts []fm.GenerationOptions) error {
	ctx := c.Request().Context()

	response, err := session.Respond(ctx, prompt, opts...)
	if err != nil {
		return h.handleError(c, err)
	}

	completionID := "chatcmpl-" + uuid.New().String()
	log.Debug("Completion response", "id", completionID, "length", len(response))

	resp := ChatCompletionResponse{
		ID:      completionID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelID,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: StringPtr(response),
				},
				FinishReason: "stop",
			},
		},
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *handler) handleStreamingResponse(c *echo.Context, session *fm.Session, prompt string, opts []fm.GenerationOptions) error {
	ctx := c.Request().Context()

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	completionID := "chatcmpl-" + uuid.New().String()
	created := time.Now().Unix()

	log.Debug("Streaming response", "id", completionID)

	roleChunk := ChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   modelID,
		Choices: []ChunkChoice{
			{
				Index: 0,
				Delta: ChunkDelta{Role: "assistant"},
			},
		},
	}
	if err := writeSSE(c, roleChunk); err != nil {
		return err
	}

	chunkCh, errCh := session.StreamResponse(ctx, prompt, opts...)

	// FM streams cumulative snapshots; convert to OpenAI-style deltas.
	var prevContent string
	for chunk := range chunkCh {
		delta := chunk
		if len(chunk) > len(prevContent) {
			delta = chunk[len(prevContent):]
		}
		prevContent = chunk

		if delta == "" {
			continue
		}

		contentChunk := ChatCompletionChunk{
			ID:      completionID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   modelID,
			Choices: []ChunkChoice{
				{
					Index: 0,
					Delta: ChunkDelta{Content: delta},
				},
			},
		}
		if err := writeSSE(c, contentChunk); err != nil {
			return err
		}
	}

	if err, ok := <-errCh; ok && err != nil {
		if err != context.Canceled {
			log.Error("Stream error", "err", err)
		}
	}

	stopReason := "stop"
	finalChunk := ChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   modelID,
		Choices: []ChunkChoice{
			{
				Index:        0,
				Delta:        ChunkDelta{},
				FinishReason: &stopReason,
			},
		},
	}
	if err := writeSSE(c, finalChunk); err != nil {
		return err
	}

	fmt.Fprint(c.Response(), "data: [DONE]\n\n")
	flush(c)

	return nil
}

func (h *handler) handleError(c *echo.Context, err error) error {
	log.Error("Generation error", "err", err)

	status := http.StatusInternalServerError
	errType := "server_error"

	switch err {
	case context.Canceled:
		return nil
	case context.DeadlineExceeded:
		status = http.StatusGatewayTimeout
		errType = "timeout_error"
	}

	return c.JSON(status, ErrorResponse{
		Error: ErrorDetail{
			Message: err.Error(),
			Type:    errType,
		},
	})
}

func buildGenerationOptions(req ChatCompletionRequest) []fm.GenerationOptions {
	hasTemp := req.Temperature != nil
	hasMax := req.MaxTokens != nil

	if !hasTemp && !hasMax {
		return nil
	}

	opts := fm.GenerationOptions{
		Temperature: -1, // default
	}
	if hasTemp {
		opts.Temperature = *req.Temperature
	}
	if hasMax {
		opts.MaxTokens = *req.MaxTokens
	}
	return []fm.GenerationOptions{opts}
}

func writeSSE(c *echo.Context, data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.Response(), "data: %s\n\n", jsonBytes)
	flush(c)
	return nil
}

func flush(c *echo.Context) {
	if f, ok := c.Response().(http.Flusher); ok {
		f.Flush()
	}
}
