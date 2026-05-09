package main

import (
	"context"
	"os"

	"github.com/Gaurav-Gosain/cider/cmd"

	fang "charm.land/fang/v2"
)

func main() {
	ctx := context.Background()
	if err := fang.Execute(ctx, cmd.Root()); err != nil {
		os.Exit(1)
	}
}
