package main

import (
	"flag"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"kardinal.kontrol-service/api"
	"kardinal.kontrol-service/database"
)

const (
	dbUsername = "postgres"
	dbHostname = "localhost"
	dbPort     = 5432
	dbName     = "kardinal"
)

func main() {
	devMode := flag.Bool("dev-mode", false, "Allow to run the service in local mode.")
	db := flag.Bool("db", false, "Enable local DB connection.")
	dbPassword := flag.String("db-password", "", "DB password.")

	flag.Parse()

	if *devMode {
		logrus.Warn("Running in dev mode. CORS fully open.")
		logrus.SetLevel(logrus.DebugLevel)
	}

	startServer(*devMode, *db, *dbPassword)
}

func startServer(isDevMode bool, isDb bool, dbPassword string) {

	var db *database.Db
	if isDb {
		dbConnectionInfo, err := database.NewDatabaseConnectionInfo(
			dbUsername,
			dbPassword,
			dbHostname,
			dbPort,
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

	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer(db)

	e := echo.New()

	if isDevMode {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
	}

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
