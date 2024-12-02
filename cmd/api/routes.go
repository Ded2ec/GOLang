package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {

	// Create a router mux
	mux := chi.NewRouter()

	//Add Middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	//Register
	mux.Get("/", app.Home)
	mux.Get("/about", app.About)
	mux.Get("/demomovies", app.AllDemoMovies)
	mux.Get("/movies", app.AllMovies)

	//Authenticate Route
	mux.Post("/authenticate", app.authenticate)
	return mux
}
