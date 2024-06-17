package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	upvotesSuffix            = ":upvotes"
	downvotesSuffix          = ":downvotes"
	validFeatureNameRegexStr = "^[a-z_]+$"
	serverPortNum            = 9111
)

var (
	validFeatureNameRegex = regexp.MustCompile(validFeatureNameRegexStr)
)

func NewVotingAppServer(ctx context.Context) (*gin.Engine, error) {
	redis, err := NewRedisConnection(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred establishing connection to redis instance")
	}

	router := gin.Default()
	router.GET("/features", getGinHandler(getFeatures, redis, ctx))
	router.POST("/upvote/:id", getGinHandler(upvoteFeature, redis, ctx))
	router.POST("/downvote/:id", getGinHandler(downvoteFeature, redis, ctx))

	return router, nil
}

func RunVotingAppServer(ctx context.Context) error {
	ginRestApiServer, err := NewVotingAppServer(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating voting app server.")
	}

	ginRestApiServer.Run(fmt.Sprintf("0.0.0.0:%v", serverPortNum))
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

// adapter to get redis connection inside api handlers
func getGinHandler(handler func(c *gin.Context, rdb *RedisConnection, ctx context.Context), redis *RedisConnection, ctxWrap context.Context) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		handler(c, redis, ctxWrap)
	}
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
