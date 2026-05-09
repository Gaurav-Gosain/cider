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
		fm.WithInstructions("You are a storyteller. Tell short, engaging stories."),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	fmt.Println("Streaming response:")
	fmt.Println("---")

	chunkCh, errCh := session.StreamResponse(
		context.Background(),
		"Tell me a very short story about a robot learning to paint.",
	)

	// Each chunk from the FM stream is a cumulative snapshot, not a delta.
	// Print only the newly-appended portion.
	var prev string
	for chunk := range chunkCh {
		if len(chunk) > len(prev) {
			fmt.Print(chunk[len(prev):])
		}
		prev = chunk
	}
	fmt.Println()
	fmt.Println("---")

	if err := <-errCh; err != nil {
		log.Fatal(err)
	}
}
