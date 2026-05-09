package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Gaurav-Gosain/cider/pkg/fm"
)

func main() {
	if err := fm.Init(); err != nil {
		log.Fatal(err)
	}

	session, err := fm.NewSession(
		fm.WithInstructions("You are a concise assistant. Answer in 1-2 sentences."),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	ctx := context.Background()

	// The session remembers context across turns.
	prompts := []string{
		"My name is Alice.",
		"What is my name?",
		"What's 2 + 2?",
		"Now multiply that result by 10.",
	}

	for _, prompt := range prompts {
		fmt.Printf("User: %s\n", prompt)

		response, err := session.Respond(ctx, prompt)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Assistant: %s\n\n", response)
	}

	// Access the full transcript
	transcript, err := session.Transcript()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transcript JSON length:", len(transcript.Raw), "bytes")
}
