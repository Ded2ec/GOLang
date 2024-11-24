package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
)

// Set Port
const port = 8080

// Appilication Struct&config Database connection

type application struct {
	DSN    string
	Domain string
	DB     *sql.DB
}

func main() {

	// Set appilication Config
	var app application
	app.Domain = "example.com"

	// Read from command line agements
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 dbname=gosampledb user=postgres password=15229 sslmode=disable timezone=UTC connect_timeout=5", "Postgres connect db")
	flag.Parse()

	// Connect to Database
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	app.DB = conn
	defer app.DB.Close()

	// http.HandleFunc("/", Hello)
	// http.HandleFunc("/about", About)
	//Start the server
	fmt.Println("Starting on", app.Domain)
	log.Printf("Starting on Port %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}

}
