package main

import (
	"log"

	api "kardinal.kloud-kontrol/api"

	"github.com/labstack/echo/v4"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()
	strictHalder := api.NewStrictHandler(server)

	e := echo.New()
	api.RegisterHandlers(e, strictHalder)
	// api.RegisterHandlers(e, server)

	// And we serve HTTP until the world ends.
	log.Fatal(e.Start("0.0.0.0:8080"))
}
