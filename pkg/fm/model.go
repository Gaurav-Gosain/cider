package fm

import (
	"runtime"
)

// ContextSize is the total context window size (input + output tokens combined)
// for Apple's on-device Foundation Model.
const ContextSize = 8192

// UseCase represents the intended use case for the language model.
type UseCase int

const (
	UseCaseGeneral        UseCase = 0
	UseCaseContentTagging UseCase = 1
)

// Guardrails represents the guardrail configuration for the language model.
type Guardrails int

const (
	GuardrailsDefault                          Guardrails = 0
	GuardrailsPermissiveContentTransformations Guardrails = 1
)

// UnavailableReason describes why the model is unavailable.
type UnavailableReason int

const (
	ReasonAppleIntelligenceNotEnabled UnavailableReason = 0
	ReasonDeviceNotEligible           UnavailableReason = 1
	ReasonModelNotReady               UnavailableReason = 2
	ReasonUnknown                     UnavailableReason = 0xFF
)

func (r UnavailableReason) String() string {
	switch r {
	case ReasonAppleIntelligenceNotEnabled:
		return "Apple Intelligence not enabled"
	case ReasonDeviceNotEligible:
		return "device not eligible"
	case ReasonModelNotReady:
		return "model not ready"
	default:
		return "unknown reason"
	}
}

// SystemLanguageModel represents a reference to Apple's on-device Foundation Model.
type SystemLanguageModel struct {
	ptr uintptr
}

// ModelOption configures a SystemLanguageModel.
type ModelOption func(*modelConfig)

type modelConfig struct {
	useCase    UseCase
	guardrails Guardrails
}

// WithUseCase sets the use case for the model.
func WithUseCase(uc UseCase) ModelOption {
	return func(c *modelConfig) {
		c.useCase = uc
	}
}

// WithGuardrails sets the guardrails for the model.
func WithGuardrails(g Guardrails) ModelOption {
	return func(c *modelConfig) {
		c.guardrails = g
	}
}

// NewSystemLanguageModel creates a new SystemLanguageModel with the given options.
func NewSystemLanguageModel(opts ...ModelOption) *SystemLanguageModel {
	cfg := &modelConfig{
		useCase:    UseCaseGeneral,
		guardrails: GuardrailsDefault,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	ptr := fmSystemLanguageModelCreate(int(cfg.useCase), int(cfg.guardrails))
	m := &SystemLanguageModel{ptr: ptr}
	runtime.SetFinalizer(m, func(m *SystemLanguageModel) { m.Close() })
	return m
}

// DefaultModel returns the default SystemLanguageModel.
func DefaultModel() *SystemLanguageModel {
	ptr := fmSystemLanguageModelGetDefault()
	m := &SystemLanguageModel{ptr: ptr}
	runtime.SetFinalizer(m, func(m *SystemLanguageModel) { m.Close() })
	return m
}

// IsAvailable checks whether the model is available for use.
func (m *SystemLanguageModel) IsAvailable() (bool, UnavailableReason) {
	var reason int
	available := fmSystemLanguageModelIsAvailable(m.ptr, &reason)
	return available, UnavailableReason(reason)
}

// Close releases the underlying C reference.
func (m *SystemLanguageModel) Close() {
	if m.ptr != 0 {
		release(m.ptr)
		m.ptr = 0
	}
}
