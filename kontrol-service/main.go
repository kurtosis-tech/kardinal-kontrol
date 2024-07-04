package main

import (
	"flag"
	"log"

	"kardinal.kontrol-service/api"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	devMode := flag.Bool("dev-mode", false, "Allow to run the service in local mode.")

	flag.Parse()

	if *devMode {
		log.Println("Running in dev mode. CORS fully open.")
	}

	startServer(*devMode)
}

func startServer(isDevMode bool) {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	e := echo.New()

	if isDevMode {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
	}

	server.RegisterExternalAndInternalApi(e)

	// And we serve HTTP until the world ends.
	log.Fatal(e.Start("0.0.0.0:8080"))
}
