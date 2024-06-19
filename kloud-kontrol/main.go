package main

import (
	"flag"
	"log"

	api "kardinal.kloud-kontrol/api"

	"github.com/labstack/echo/v4"
)

func main() {
	applyLocally := flag.Bool("apply-directly", false, "Apply changes directly to configured k8s")

	flag.Parse()

	if *applyLocally {
		log.Println("Applying changes directly.")
	} else {
		log.Println("Server configuration for pulling.")
	}

	startServer(*applyLocally)
}

func startServer(applyLocally bool) {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()
	strictHalder := api.NewStrictHandler(server)

	e := echo.New()
	api.RegisterHandlers(e, strictHalder)
	// api.RegisterHandlers(e, server)

	// And we serve HTTP until the world ends.
	log.Fatal(e.Start("0.0.0.0:8080"))
}
