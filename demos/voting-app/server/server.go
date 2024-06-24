package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	upvotesSuffix            = ":upvotes"
	downvotesSuffix          = ":downvotes"
	validFeatureNameRegexStr = "^[a-z_]+$"
	serverPortNumStr         = "9111"
	serverStopGracePeriod    = 1 * time.Minute
)

var (
	validFeatureNameRegex = regexp.MustCompile(validFeatureNameRegexStr)
)

func NewVotingAppServer(ctx context.Context) (http.Handler, *logrus.Logger, error) {
	redis, err := NewRedisConnection(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred establishing connection to redis instance")
	}

	router := gin.Default()

	logger := setupLogrus()

	router.Use(getLoggingMiddleware(logger))

	router.GET("/features", getGinHandler(getFeatures, redis, ctx))
	router.POST("/upvote/:id", getGinHandler(upvoteFeature, redis, ctx))
	router.POST("/downvote/:id", getGinHandler(downvoteFeature, redis, ctx))

	return router.Handler(), logger, nil
}

// adapter to get redis connection inside api handlers
func getGinHandler(handler func(c *gin.Context, rdb *RedisConnection, ctx context.Context), redis *RedisConnection, ctxWrap context.Context) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		handler(c, redis, ctxWrap)
	}
}

func getLoggingMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		entry := log.WithFields(logrus.Fields{
			"status":  c.Writer.Status(),
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"ip":      c.ClientIP(),
			"latency": duration,
		})
		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			entry.Infof("request completed")
		}
	}
}

func setupLogrus() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
	return log
}

func RunVotingAppServerUntilInterrupted(ctx context.Context) error {
	handler, logger, err := NewVotingAppServer(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating voting app server.")
	}

	srv := &http.Server{
		Addr:     net.JoinHostPort("0.0.0.0", serverPortNumStr),
		Handler:  handler,
		ErrorLog: log.New(logrus.StandardLogger().Out, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
	if err = runServerUntilInterrupted(ctx, srv, logger); err != nil {
		return stacktrace.Propagate(err, "An error occurred while running server.")
	}
	return nil
}

func runServerUntilInterrupted(ctx context.Context, srv *http.Server, logger *logrus.Logger) error {
	termSignalChan := make(chan os.Signal, 1)
	signal.Notify(termSignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverStopChan := make(chan struct{}, 1)
	go func() {
		<-termSignalChan
		interruptSignal := struct{}{}
		serverStopChan <- interruptSignal
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("An error occurred starting server:\n%v", err)
		}
	}()

	<-serverStopChan
	serverStoppedChan := make(chan interface{})
	go func() {
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatalf("An error occurred shutting down server:\n%v", err)
		}
		serverStoppedChan <- nil
	}()
	select {
	case <-serverStoppedChan:
		logger.Debugf("Voting App Server exited gracefully.")
	case <-time.After(serverStopGracePeriod):
		if err := srv.Close(); err != nil {
			logger.Infof("An error occurred forcefully closing the server:\n%v", err)
		}
	}
	return nil
}

type Feature struct {
	Name      string `json:"name"`
	Upvotes   int    `json:"upvotes"`
	Downvotes int    `json:"downvotes"`
}

func upvoteFeature(c *gin.Context, redisClient *RedisConnection, ctx context.Context) {
	featureName := c.Param("id")
	if !isValidFeatureName(featureName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature name"})
		return
	}

	logrus.Infof("Upvoting feature: %s", featureName)

	err := createFeatureIdempotently(ctx, redisClient, featureName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
		return
	}

	_, err = redisClient.rdb.Incr(ctx, getUpvoteKey(featureName)).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upvote feature", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully upvoted feature"})
}

func downvoteFeature(c *gin.Context, redisClient *RedisConnection, ctx context.Context) {
	featureName := c.Param("id")
	if !isValidFeatureName(featureName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature name"})
		return
	}

	logrus.Infof("Downvoting feature: %s", featureName)

	err := createFeatureIdempotently(ctx, redisClient, featureName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
		return
	}

	_, err = redisClient.rdb.Incr(ctx, getDownvoteKey(featureName)).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to downvote feature", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully downvoted feature"})
}

func getFeatures(c *gin.Context, redisClient *RedisConnection, ctx context.Context) {
	var cursor uint64
	keys := make([]string, 0)
	for {
		var scanKeys []string
		var err error
		// assume for each upvote key, there is a downvote one, so only scan for upvotes
		scanKeys, cursor, err = redisClient.rdb.Scan(ctx, cursor, fmt.Sprintf("*%v", upvotesSuffix), 10).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
			return
		}
		keys = append(keys, scanKeys...)
		if cursor == 0 {
			break
		}
	}

	features := []Feature{}
	for _, key := range keys {
		featureName := getFeatureNameFromKey(key)
		upvotesStr, err := redisClient.rdb.Get(ctx, getUpvoteKey(featureName)).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
			return
		}
		upvotes, err := strconv.Atoi(upvotesStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
			return
		}

		downvotesStr, err := redisClient.rdb.Get(ctx, getDownvoteKey(featureName)).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
			return
		}
		downvotes, err := strconv.Atoi(downvotesStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
			return
		}

		features = append(features, Feature{
			Name:      featureName,
			Upvotes:   upvotes,
			Downvotes: downvotes,
		})
	}

	c.JSON(http.StatusOK, features)
}

func createFeatureIdempotently(ctx context.Context, rdb *RedisConnection, featureName string) error {
	upvoteKey := getUpvoteKey(featureName)
	downvoteKey := getDownvoteKey(featureName)

	tx := rdb.rdb.TxPipeline()
	upvoteExists := tx.Exists(ctx, getUpvoteKey(featureName))
	downvoteExists := tx.Exists(ctx, getDownvoteKey(featureName))
	_, err := tx.Exec(ctx)
	if err != nil {
		return err
	}

	doesFeatureExist := upvoteExists.Val() == 1 && downvoteExists.Val() == 1
	if !doesFeatureExist {
		rdb.rdb.Set(ctx, upvoteKey, 0, 0)
		rdb.rdb.Set(ctx, downvoteKey, 0, 0)
	}

	return nil
}

func getUpvoteKey(featureName string) string {
	return fmt.Sprintf("%v:%v", featureName, upvotesSuffix)
}

func getDownvoteKey(featureName string) string {
	return fmt.Sprintf("%v:%v", featureName, downvotesSuffix)
}

func getFeatureNameFromKey(key string) string {
	return strings.Split(key, ":")[0]
}

func isValidFeatureName(featureName string) bool {
	return validFeatureNameRegex.MatchString(featureName)
}

func persistRedis(rdb *RedisConnection) {
	// TODO
}
