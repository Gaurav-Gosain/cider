package fm

import (
	"context"
)

// Extract sends a prompt and unmarshals the structured response directly into a typed value.
// It combines SchemaFor, RespondWithSchema, and Unmarshal into a single call.
//
// Example:
//
//	type Sentiment struct {
//	    Score float64 `json:"score" description:"Sentiment score from -1.0 to 1.0"`
//	    Label string  `json:"label" description:"Sentiment label" enum:"positive,negative,neutral"`
//	}
//
//	var result Sentiment
//	err := fm.Extract(ctx, session, "Analyze: I love this product!", &result)
func Extract[T any](ctx context.Context, session *Session, prompt string, dest *T, opts ...GenerationOptions) error {
	schema := SchemaFor[T]()

	content, err := session.RespondWithSchema(ctx, prompt, schema)
	if err != nil {
		return err
	}
	defer content.Close()

	return Unmarshal(content, dest)
}
