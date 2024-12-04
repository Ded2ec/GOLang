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
	mux.Get("/refresh", app.refreshToken)
	mux.Get("/logout", app.logout)

	mux.Get("/movies", app.AllMovies)
	mux.Get("/movies/{id}", app.GetMovie)
	mux.Get("/genres", app.AllGenres)

	mux.Route("/admin", func(mux chi.Router) {
		// Protected routes
		mux.Use(app.authRequired)

		mux.Get("/movies", app.MovieCatalog)
		mux.Get("/movies/{id}", app.MovieForEdit)
		mux.Post("/movies", app.InsertMovie)
		mux.Put("/movies/{id}", app.UpdateMovie)
		mux.Delete("/movies/{id}", app.DeleteMovie)
	})

	return mux
}
