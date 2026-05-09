package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Gaurav-Gosain/cider/pkg/fm"
)

// CalculatorArgs defines the arguments for the calculator tool.
type CalculatorArgs struct {
	Operation string  `json:"operation" description:"The arithmetic operation" enum:"add,subtract,multiply,divide"`
	A         float64 `json:"a"         description:"First number"`
	B         float64 `json:"b"         description:"Second number"`
}

func main() {
	if err := fm.Init(); err != nil {
		log.Fatal(err)
	}

	calc := fm.FuncTool("calculator", "Performs basic arithmetic operations", func(args CalculatorArgs) (string, error) {
		var result float64
		switch strings.ToLower(args.Operation) {
		case "add":
			result = args.A + args.B
		case "subtract":
			result = args.A - args.B
		case "multiply":
			result = args.A * args.B
		case "divide":
			if args.B == 0 {
				return "Error: division by zero", nil
			}
			result = args.A / args.B
		default:
			return fmt.Sprintf("Unknown operation: %s", args.Operation), nil
		}
		return fmt.Sprintf("%.2f", result), nil
	})

	session, err := fm.NewSession(
		fm.WithInstructions("You are a helpful math assistant. Use the calculator tool for arithmetic."),
		fm.WithTools(calc),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	response, err := session.Respond(
		context.Background(),
		"What is 42 multiplied by 17?",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", response)
}
