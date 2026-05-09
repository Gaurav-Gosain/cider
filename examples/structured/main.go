package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Gaurav-Gosain/cider/pkg/fm"
)

// Person represents a person's profile for structured extraction.
type Person struct {
	Name       string `json:"name"       description:"The person's full name"`
	Age        int    `json:"age"        description:"The person's age"`
	Occupation string `json:"occupation" description:"The person's job title"`
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

	schema := fm.SchemaFor[Person]()

	content, err := session.RespondWithSchema(
		context.Background(),
		"Create a fictional character who is a marine biologist",
		schema,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer content.Close()

	var person Person
	if err := fm.Unmarshal(content, &person); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name:       %s\n", person.Name)
	fmt.Printf("Age:        %d\n", person.Age)
	fmt.Printf("Occupation: %s\n", person.Occupation)
}
