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

	model := fm.DefaultModel()
	defer model.Close()

	available, reason := model.IsAvailable()
	if !available {
		log.Fatalf("Model not available: %s", reason)
	}
	fmt.Println("Model is available.")

	session, err := fm.NewSession(
		fm.WithInstructions("You are a helpful assistant. Keep responses concise."),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	response, err := session.Respond(context.Background(), "What is the capital of France?")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", response)
}
