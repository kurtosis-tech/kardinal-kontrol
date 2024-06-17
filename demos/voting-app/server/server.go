package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
)

const (
	upvotesSuffix            = ":upvotes"
	downvotesSuffix          = ":downvotes"
	validFeatureNameRegexStr = "^[a-z]+$"
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
	ctx, cancel := signal.NotifyContext(ctx)
	defer cancel()
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

func upvoteFeature(c *gin.Context, rdb *RedisConnection, ctx context.Context) {
	featureName := c.Param(":id")
	if !isValidFeatureName(featureName) {
		c.String(400, "invalid feature name.")
	}
	logrus.Infof("upvoting %v", featureName)

	err := createFeatureIdempotently(ctx, rdb, featureName)
	if err != nil {
		c.String(400, "An error occurred checking if feature exists/creating it: %v", err.Error())
		return
	}

	_, err = rdb.rdb.Incr(ctx, getUpvoteKey(featureName)).Result()
	if err != nil {
		c.String(400, "An error occurred upvoting feature: %v\n%v", featureName, err.Error())
		return
	}

	c.Data(http.StatusOK, "application/json", []byte("successfully upvoted feature."))
}

func downvoteFeature(c *gin.Context, rdb *RedisConnection, ctx context.Context) {
	featureName := c.Param(":id")
	if !isValidFeatureName(featureName) {
		c.String(400, "invalid feature name.")
	}
	logrus.Infof("downvoting %v", featureName)

	err := createFeatureIdempotently(ctx, rdb, featureName)
	if err != nil {
		c.String(400, "An error occurred checking if feature exists/creating it: %v", err.Error())
	}

	_, err = rdb.rdb.Incr(ctx, getDownvoteKey(featureName)).Result()
	if err != nil {
		c.String(400, "An error occurred downvoting feature: %v\n%v", featureName, err.Error())
	}

	c.Data(http.StatusOK, "application/json", []byte("successfully downvoted feature."))
}

func getFeatures(c *gin.Context, rdb *RedisConnection, ctx context.Context) {
	var cursor uint64
	keys := make([]string, 0)
	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = rdb.rdb.Scan(ctx, cursor, "*", 10).Result()
		if err != nil {
			c.String(400, "An error occurred scanning keys from Redis to get features:\n%v", err.Error())
		}
		keys = append(keys, scanKeys...)
		if cursor == 0 {
			break
		}
	}

	features := map[string]Feature{}
	for _, key := range keys {
		featureName := getFeatureNameFromKey(key)
		if _, ok := features[featureName]; ok {
			// if we've already seen the feature don't need to retrieve again
			continue
		}
		upvotesStr, err := rdb.rdb.Get(ctx, getUpvoteKey(featureName)).Result()
		if err != nil {
			c.String(400, "An error occurred scanning keys from Redis to get features.")
		}
		upvotes, err := strconv.Atoi(upvotesStr)
		if err != nil {
			c.String(400, "An error occurred converting upvote string to int")
		}
		downvotesStr, err := rdb.rdb.Get(ctx, getDownvoteKey(featureName)).Result()
		if err != nil {
			c.String(400, "An error occurred scanning keys from Redis to get features.")
		}
		downvotes, err := strconv.Atoi(downvotesStr)
		if err != nil {
			c.String(400, "An error occurred converting downvote string to int")
		}
		features[featureName] = Feature{
			Name:      featureName,
			Upvotes:   upvotes,
			Downvotes: downvotes,
		}
	}

	c.JSONP(http.StatusOK, features)
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
	logrus.Infof("feature name: %v", strings.Split(key, ":"))
	return strings.Split(key, ":")[0]
}

func isValidFeatureName(featureName string) bool {
	return validFeatureNameRegex.Match([]byte(featureName))
}

func persistRedis(rdb *RedisConnection) {
	// TODO
}
