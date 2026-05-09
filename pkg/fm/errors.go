package fm

import (
	"errors"
	"fmt"
)

// ErrorCode represents error codes returned by the Foundation Models C API.
type ErrorCode int

const (
	ErrorCodeSuccess                    ErrorCode = 0
	ErrorCodeExceededContextWindowSize  ErrorCode = 1
	ErrorCodeAssetsUnavailable          ErrorCode = 2
	ErrorCodeGuardrailViolation         ErrorCode = 3
	ErrorCodeUnsupportedGuide           ErrorCode = 4
	ErrorCodeUnsupportedLanguageOrLocal ErrorCode = 5
	ErrorCodeDecodingFailure            ErrorCode = 6
	ErrorCodeRateLimited                ErrorCode = 7
	ErrorCodeConcurrentRequests         ErrorCode = 8
	ErrorCodeRefusal                    ErrorCode = 9
	ErrorCodeInvalidSchema              ErrorCode = 10
	ErrorCodeUnknown                    ErrorCode = 255
)

// FoundationModelsError is the base error type for all Foundation Models errors.
type FoundationModelsError struct {
	Code    ErrorCode
	Message string
}

func (e *FoundationModelsError) Error() string {
	return fmt.Sprintf("foundation models error (code %d): %s", e.Code, e.Message)
}

var (
	ErrExceededContextWindowSize = errors.New("exceeded context window size")
	ErrAssetsUnavailable         = errors.New("assets unavailable")
	ErrGuardrailViolation        = errors.New("guardrail violation")
	ErrUnsupportedGuide          = errors.New("unsupported guide")
	ErrUnsupportedLanguage       = errors.New("unsupported language or locale")
	ErrDecodingFailure           = errors.New("decoding failure")
	ErrRateLimited               = errors.New("rate limited")
	ErrConcurrentRequests        = errors.New("concurrent requests")
	ErrRefusal                   = errors.New("refusal")
	ErrInvalidSchema             = errors.New("invalid generation schema")
	ErrUnknown                   = errors.New("unknown error")
	ErrModelUnavailable          = errors.New("model unavailable")
	ErrCancelled                 = errors.New("request cancelled")
)

// errorFromCode creates a Go error from a C error code and optional description.
func errorFromCode(code int, desc string) error {
	ec := ErrorCode(code)
	if ec == ErrorCodeSuccess {
		return nil
	}

	base := &FoundationModelsError{Code: ec, Message: desc}

	switch ec {
	case ErrorCodeExceededContextWindowSize:
		return fmt.Errorf("%w: %s", ErrExceededContextWindowSize, base)
	case ErrorCodeAssetsUnavailable:
		return fmt.Errorf("%w: %s", ErrAssetsUnavailable, base)
	case ErrorCodeGuardrailViolation:
		return fmt.Errorf("%w: %s", ErrGuardrailViolation, base)
	case ErrorCodeUnsupportedGuide:
		return fmt.Errorf("%w: %s", ErrUnsupportedGuide, base)
	case ErrorCodeUnsupportedLanguageOrLocal:
		return fmt.Errorf("%w: %s", ErrUnsupportedLanguage, base)
	case ErrorCodeDecodingFailure:
		return fmt.Errorf("%w: %s", ErrDecodingFailure, base)
	case ErrorCodeRateLimited:
		return fmt.Errorf("%w: %s", ErrRateLimited, base)
	case ErrorCodeConcurrentRequests:
		return fmt.Errorf("%w: %s", ErrConcurrentRequests, base)
	case ErrorCodeRefusal:
		return fmt.Errorf("%w: %s", ErrRefusal, base)
	case ErrorCodeInvalidSchema:
		return fmt.Errorf("%w: %s", ErrInvalidSchema, base)
	default:
		return fmt.Errorf("%w: %s", ErrUnknown, base)
	}
}

// extractError extracts an error from C out-parameters, freeing the description string.
func extractError(errorCode int, errorDescPtr uintptr) error {
	if errorCode == 0 {
		freeStr(errorDescPtr)
		return nil
	}
	desc := ""
	if errorDescPtr != 0 {
		desc = goString(errorDescPtr)
		freeStr(errorDescPtr)
	}
	return errorFromCode(errorCode, desc)
}

// statusCodeToError converts a callback status code to an error.
func statusCodeToError(status int, content string) error {
	if status == 0 {
		return nil
	}
	return errorFromCode(status, content)
}
