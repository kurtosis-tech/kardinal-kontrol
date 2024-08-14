package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"kardinal.kontrol-service/api"
	"kardinal.kontrol-service/database"
)

func main() {
	devMode := flag.Bool("dev-mode", false, "Allow to run the service in local mode.")

	flag.Parse()

	if *devMode {
		logrus.Warn("Running in dev mode. CORS fully open.")
		logrus.SetLevel(logrus.DebugLevel)
	}

	startServer(*devMode)
}

func startServer(isDevMode bool) {

	var db *database.Db
	dbHostname := os.Getenv("DB_HOSTNAME")
	if dbHostname != "" {
		dbUsername := os.Getenv("DB_USERNAME")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
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
	}

	// Create a new Segment analytics client instance.
	// analyticsClient is not initialized in dev mode so events are not reported to Segment
	analyticsWrapper := api.NewAnalyticsWrapper(isDevMode)
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

	server.RegisterExternalAndInternalApi(e)

	// And we serve HTTP until the world ends.
	logrus.Fatal(e.Start("0.0.0.0:8080"))
}
