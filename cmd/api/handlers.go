package main

import (
	"backend/internal/models"
	"encoding/json"
	"errors"
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

// Funtion Authentication and Create Tokenpairs

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// read json payload (อ่านข้อมูล JSON ที่ส่งมา)
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate user against database (ตรวจสอบข้อมูลผู้ใช้จากฐานข้อมูล)
	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}
	// check password against hash (ตรวจสอบรหัสผ่าน)
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}
	// create a jwt user (สร้าง jwt user)
	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	// generate tokens (สร้างโทเคน)
	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	app.writeJSON(w, http.StatusAccepted, tokens)

}
