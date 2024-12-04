package main

import (
	"backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
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

// ฟังก์ชันสำหรับแสดงรายละเอียดหนังตาม ID
func (app *application) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movie, err := app.DB.OneMovie(movieID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movie)
}

// ฟังก์ชันดึงข้อมูลหนังและประเภทหนังสำหรับการแก้ไข
func (app *application) MovieForEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movie, genres, err := app.DB.OneMovieForEdit(movieID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload = struct {
		Movie  *models.Movie   `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		movie,
		genres,
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
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

// ฟังก์ชันสำหรับการ Refresh Token
func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			// parse the token to get the claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get the user id from the token claims
			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.DB.GetUserByID(userID)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			tokenPairs, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJSON(w, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

			app.writeJSON(w, http.StatusOK, tokenPairs)

		}
	}
}

// ฟังก์ชันสำหรับการ Logout
func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}
