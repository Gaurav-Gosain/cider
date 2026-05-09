package fm

import (
	"encoding/json"
)

// Transcript represents the conversation history of a session.
type Transcript struct {
	Raw string
}

// GetTranscript retrieves the transcript from a session.
func (s *Session) Transcript() (*Transcript, error) {
	var errorCode int
	var errorDesc uintptr

	cs := fmLanguageModelSessionGetTranscriptJSONString(s.ptr, &errorCode, &errorDesc)
	if err := extractError(errorCode, errorDesc); err != nil {
		return nil, err
	}
	if cs == 0 {
		return &Transcript{Raw: "{}"}, nil
	}
	defer freeStr(cs)

	return &Transcript{Raw: goString(cs)}, nil
}

// ToMap parses the transcript JSON into a map.
func (t *Transcript) ToMap() (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(t.Raw), &result)
	return result, err
}
