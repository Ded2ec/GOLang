package main

import (
	"fmt"
	"log"
	"net/http"
)

// Set Port
const port = 8080

// Appilication Struct&config Database connection

type application struct {
	Domain string
}

func main() {

	// Set appilication Config

	var app application
	app.Domain = "example.com"
	// Read from command line agements

	// Connect to Database
	// http.HandleFunc("/", Hello)
	// http.HandleFunc("/about", About)
	//Start the server
	fmt.Println("Starting on", app.Domain)
	log.Printf("Starting on Port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}

}
