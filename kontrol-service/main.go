package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	cli_api "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"kardinal.kontrol-service/api"
	"kardinal.kontrol-service/database"
)

func main() {
	var devMode bool
	devMode = *flag.Bool("dev-mode", false, "Allow to run the service in local mode.")
	if !devMode {
		devModeEnvVarStr := os.Getenv("DEV_MODE")
		if devModeEnvVarStr == "true" {
			devMode = true
		}
	}

	flag.Parse()

	if devMode {
		logrus.Warn("Running in dev mode. CORS fully open.")
		logrus.SetLevel(logrus.DebugLevel)
	}

	startServer(devMode)
}

func startServer(isDevMode bool) {

	dbHostname := os.Getenv("DB_HOSTNAME")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if dbHostname == "" || dbUsername == "" || dbPassword == "" || dbName == "" {
		logrus.Fatal("One of the following environment variables is not set: DB_HOSTNAME, DB_USERNAME, DB_PASSWORD, DB_NAME")
	}
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		logrus.Fatal("An error occurred parsing the DB port number", err)
	}

	dbConnectionInfo, err := database.NewDatabaseConnectionInfo(
		dbUsername,
		dbPassword,
		dbHostname,
		uint16(dbPort),
		dbName,
	)
	if err != nil {
		logrus.Fatal("An error occurred creating a database connection configuration based on the input provided", err)
	}

	db, err := database.NewDb(dbConnectionInfo)
	if err != nil {
		logrus.Fatal("An error occurred creating the db connection", err)
	}

	err = db.Migrate()
	if err != nil {
		logrus.Fatal("An error occurred migrating the DB", err)
	}

	// Create a new Segment analytics client instance.
	// analyticsClient is not initialized in dev mode so events are not reported to Segment
	analyticsWriteKeyEnvVarStr := os.Getenv("ANALYTICS_WRITE_KEY")
	analyticsWrapper := api.NewAnalyticsWrapper(isDevMode, analyticsWriteKeyEnvVarStr)
	defer analyticsWrapper.Close()

	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer(db, analyticsWrapper)

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			logrus.WithFields(logrus.Fields{
				"URI":    values.URI,
				"status": values.Status,
			}).Info("request")
			return nil
		},
	}))

	// Panic handler
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {

					msg := "The server could not handle this request.  Please make sure you are using the latest Kardinal API.  For the CLI, it means using the latest CLI release.  You can open an issue against the Kardinal repo if you continue to get this error at https://github.com/kurtosis-tech/kardinal/issues/new"
					internalServerErrorResponse := cli_api.ErrorJSONResponse{
						Error: "Internal Server Error",
						Msg:   &msg,
					}

					// Handle the panic and return a 500 error response
					c.JSON(http.StatusInternalServerError, internalServerErrorResponse)
					var debugMsg string
					switch recoverErr := r.(type) {
					case string:
						debugMsg = recoverErr
					case error:
						debugMsg = recoverErr.Error()
					default:
						debugMsg = "recover didn't get error msg"
					}
					logrus.Errorf("HTTP server handled this internal panic: %s", debugMsg)
				}
			}()
			return next(c)
		}
	})

	server.RegisterExternalAndInternalApi(e)

	// And we serve HTTP until the world ends.
	logrus.Fatal(e.Start("0.0.0.0:8080"))
}
