package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Gaurav-Gosain/cider/pkg/fm"
)

// Sentiment represents the sentiment analysis result.
type Sentiment struct {
	Score float64 `json:"score" description:"Sentiment score from -1.0 to 1.0"`
	Label string  `json:"label" description:"Sentiment label" enum:"positive,negative,neutral"`
}

func main() {
	if err := fm.Init(); err != nil {
		log.Fatal(err)
	}

	session, err := fm.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	var result Sentiment
	if err := fm.Extract(context.Background(), session, "Analyze the sentiment: I absolutely love this product, it changed my life!", &result); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Label: %s\n", result.Label)
	fmt.Printf("Score: %.2f\n", result.Score)
}
