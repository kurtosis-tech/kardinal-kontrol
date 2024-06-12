package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"os/signal"
)

type VotingAppServer struct {
	// connection to redis
	redis *RedisConnection
}

func NewVotingAppServer() *VotingAppServer {
	// pick up basic config from the environment variables
	// configure middleware here
	// add all the routes here
	return &VotingAppServer{}
}

func RunVotingAppServer(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx)
	defer cancel()
	// create voting app server
	return nil
}

type Feature struct {
	Id          string
	Name        string
	Description string
	Upvotes     int
	Downvotes   int
}

func createFeature(c gin.Context) error {
	return nil
}

func upvoteFeature(c gin.Context) error {
	return nil
}

func downvoteFeature(ctx gin.Context) error {
	return nil
}
