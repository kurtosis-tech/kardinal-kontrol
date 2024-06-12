package main

import (
	"context"
	"fmt"
	"os"
	"votingapp/server"
)

func main() {
	ctx := context.Background()
	if err := server.RunVotingAppServer(ctx); err != nil {
		fmt.Errorf("An error occurred running voting app server: %v", err)
		os.Exit(1)
	}
	// 1. figure out what rest api framework to use, use a very lightweight one for the sake of this, one with not a lot of boilerplate
	// what do we need for the server?
	// we could create an http server manually
	// create a handler
	// register middleware on that handler
	// create an http server using that handler that listen on a port
	// start server here
}
