package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"os/signal"
)

func NewVotingAppServer() (http.Handler, *gin.Engine, error) {
	// create connection to Redis db
	redis, err := NewRedisConnection()
	if err != nil {
		return nil, nil, nil
	}

	// setup rest api server with gin
	router := gin.Default()
	router.POST("/createFeatures", getGinHandler(createFeature, redis))
	router.POST("/upvoteFeature", getGinHandler(upvoteFeature, redis))
	router.POST("/downvoteFeature", getGinHandler(downvoteFeature, redis))

	return router.Handler(), router, nil
}

func RunVotingAppServer(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx)
	defer cancel()

	_, ginRestApiServer, err := NewVotingAppServer()
	if err != nil {
		return err
	}
	ginRestApiServer.Run("0.0.0.0:9000")

	// TODO: switch to graceful shutdown
	//httpServer := &http.Server{
	//	Addr:    net.JoinHostPort("0.0.0.0", "9000"),
	//	Handler: handler,
	//}
	//go func() {
	//	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	//		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	//	}
	//}()
	//var wg sync.WaitGroup
	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	<-ctx.Done()
	//	// make a new context for the Shutdown (thanks Alessandro Rosetti)
	//	shutdownCtx := context.Background()
	//	shutdownCtx, cancel := context.WithTimeout(ctx, 10 * time.Second)
	//	defer cancel()
	//	if err := httpServer.Shutdown(shutdownCtx); err != nil {
	//		fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
	//	}
	//}()
	//wg.Wait()
	return nil
}

type Feature struct {
	Name        string `json:"id"`
	Description string `json:"id"`
}

type Upvote struct {
	FeatureName string `josn:"feature"`
}

type Downvote struct {
	FeatureName string `josn:"feature"`
}

func createFeature(ctx *gin.Context, db *RedisConnection) {
	var feature Feature
	if err := ctx.BindJSON(&feature); err != nil {
		return
	}

	// put it in the redis db

	ctx.Data(http.StatusOK, "application/json", []byte("successfully created your feature."))
}

func upvoteFeature(ctx *gin.Context, db *RedisConnection) {
	var upvote Upvote
	if err := ctx.BindJSON(&upvote); err != nil {
		return
	}

	// update the redis db

	ctx.Data(http.StatusOK, "application/json", []byte("successfully upvoted feature."))
}

func downvoteFeature(ctx *gin.Context, db *RedisConnection) {
	var downvote Downvote
	if err := ctx.BindJSON(&downvote); err != nil {
		return
	}

	// update the redis db

	ctx.Data(http.StatusOK, "application/json", []byte("successfully downvoted feature."))
}

// adapter to get redis connection inside api handlers
func getGinHandler(handler func(ctx *gin.Context, db *RedisConnection), redis *RedisConnection) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		handler(ctx, redis)
	}
}
