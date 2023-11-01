package main

import (
	"log"
	"github.com/kendax/calculator_go_internal/routes"
	"os"
)

func main() {
	//Detect the port from the environment variable and if not set default to port 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	//Initialize and setup routing for the web application
	r := routes.SetupRoutes()

	log.Println("Listening on https://localhost:3000")

	//Start a HTTP server and make it listen for incoming requests on the defined port
	r.Run(":" + port)
}