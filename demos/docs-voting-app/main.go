package main

import (
	"context"
	"fmt"
	"os"
	"votingapp/server"
)

func main() {
	ctx := context.Background()
	if err := server.RunVotingAppServerUntilInterrupted(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred running voting app server: %v", err)
		os.Exit(1)
	}
}
