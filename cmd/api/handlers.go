package main

import (
	"backend/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "Hello, world! ", app.Domain)
	// Json Data
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Go Movies up and running",
		Version: "1.0.0",
	}

	// out, err := json.Marshal(payload)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// w.Write(out)

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) About(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "About, Nontachai")
}

// TestData Function All name movie

func (app *application) AllDemoMovies(w http.ResponseWriter, r *http.Request) {

	// create movie
	var movies []models.Movie

	// set yyyy-mm-dd
	rd, _ := time.Parse("2006-01-02", "1981-06-07")

	//Data Movie 1
	yellowboy := models.Movie{
		ID:          1,
		Title:       "Yellowboy",
		ReleaseDate: rd,
		MPAARating:  "20+",
		RunTime:     116,
		Description: "A very nice movie",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// Data Movie  slice
	movies = append(movies, yellowboy)

	rd, _ = time.Parse("2006-01-02", "1982-06-07")

	//Data Movie 2
	swingingboy := models.Movie{
		ID:          2,
		Title:       "Swingingboy",
		ReleaseDate: rd,
		MPAARating:  "20+",
		RunTime:     116,
		Description: "A very nice movie",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// Data Movie  slice
	movies = append(movies, swingingboy)

	out, err := json.Marshal(movies)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)

}

// Data  All  movie
func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	//All movie from app.DB
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, movies)
}
