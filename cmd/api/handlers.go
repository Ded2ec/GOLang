package main

import (
	"backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// UserRegisterPayload is the request payload for user login
// swagger:parameters login
type UserLoginPayload struct {
	// Required: true
	// Example: "
	Email string `json:"email"`
	// Required: true
	// Example: "password123"
	Password string `json:"password"`
}

// UserRegisterPayload is the request payload for user registration
// swagger:parameters register
type UserRegisterPayload struct {
	// Required: true
	// Example: "John"
	FirstName string `json:"first_name"`
	// Required: true
	// Example: "Doe"
	LastName string `json:"last_name"`
	// Required: true
	// Example: "john@example.com"
	Email string `json:"email"`
	// Required: true
	// Example: "password123"
	Password string `json:"password"`
}

// login ทำการ login และสร้าง TokenPairs
// @Summary Authentication และสร้าง TokenPairs
// @Description รับข้อมูลอีเมลและรหัสผ่านของผู้ใช้และตรวจสอบความถูกต้อง หลังจากนั้นสร้าง JWT TokenPairs
// @Tags Authentication
// @Accept json
// @Produce json
// @Param requestPayload body UserLoginPayload true "User credentials" example({"email": "string", "password": "string"})
// @Success 202 {object} map[string]interface{} "Token pairs" example({"access_token": "string", "refresh_token": "string"})
// @Failure 400 {object} map[string]interface{} "Bad Request" example({"error": "Bad Request"})
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error": "Internal Server Error"})
// @Router /api/v1/login [post]
func (app *application) login(w http.ResponseWriter, r *http.Request) {
	// read json payload (อ่านข้อมูล JSON ที่ส่งมา)
	var requestPayload UserLoginPayload

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

	// set refresh token cookie (ตั้งค่า cookie สำหรับ refresh token)
	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	// create the response payload (สร้าง payload สำหรับ response)
	responsePayload := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		} `json:"user"`
	}{
		AccessToken:  tokens.Token, // แก้ไขให้ตรงกับฟิลด์ Token ของคุณ
		RefreshToken: tokens.RefreshToken,
		User: struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		}{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}

	// write the response as JSON (เขียน response เป็น JSON)
	app.writeJSON(w, http.StatusAccepted, responsePayload)
}

// register เพิ่มผู้ใช้ใหม่ในระบบ
// @Summary เพิ่มผู้ใช้ใหม่
// @Description รับข้อมูลผู้ใช้ใหม่และบันทึกลงในระบบ
// @Tags Authentication
// @Accept json
// @Produce json
// @Param requestPayload body UserRegisterPayload true "User registration data" example({"first_name": "John", "last_name": "Doe", "email": "john@example.com", "password": "password123"})
// @Success 201 {object} map[string]string "message" example({"message": "User created"})
// @Failure 400 {object} map[string]string "Bad Request" example({"error": "Bad Request"})
// @Failure 500 {object} map[string]string "Internal Server Error" example({"error": "Internal Server Error"})
// @Router /api/v1/register [post]
func (app *application) register(w http.ResponseWriter, r *http.Request) {
	var requestPayload UserRegisterPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// ตรวจสอบว่าอีเมลนี้มีอยู่แล้วในระบบหรือไม่
	existingUser, _ := app.DB.GetUserByEmail(requestPayload.Email)
	if existingUser != nil {
		app.errorJSON(w, errors.New("email already exists"), http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), bcrypt.DefaultCost)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// Create new user
	user := models.User{
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Email:     requestPayload.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert user to database
	_, err = app.DB.InsertUser(user)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "User created",
	}

	app.writeJSON(w, http.StatusCreated, resp)
}

// refreshToken รีเฟรชโทเคน JWT
// @Summary รีเฟรชโทเคน JWT
// @Description ตรวจสอบโทเคนที่หมดอายุและสร้างโทเคนใหม่สำหรับผู้ใช้
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]string "Token pairs" example({"access_token": "string", "refresh_token": "string"})
// @Failure 401 {object} map[string]string "Unauthorized" example({"error": "Unauthorized"})
// @Failure 500 {object} map[string]string "Internal Server Error" example({"error": "Internal Server Error"})
// @Router /api/v1/refresh [get]
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

// logout ออกจากระบบ
// @Summary ออกจากระบบ
// @Description ลบโทเคนรีเฟรชของผู้ใช้ออกจากระบบ
// @Tags Authentication
// @Produce json
// @Success 202 {object} map[string]string "Accepted" example({"message": "Accepted"})
// @Failure 500 {object} map[string]string "Internal Server Error" example({"error": "Internal Server Error"})
// @Router /api/v1/logout [get]
func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

// Home ฟังก์ชันสำหรับตรวจสอบสถานะการทำงานของ API
// @Summary ตรวจสอบสถานะการทำงานของ API
// @Description แสดงสถานะและข้อมูลเกี่ยวกับ API
// @Tags Home
// @Produce json
// @Success 200 {object} map[string]interface{} "{"status":"active","message":"Go Movies up and running","version":"1.0.0"}"
// @Router /api/v1/ [get]
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

// About ฟังก์ชันสำหรับแสดงข้อมูลเกี่ยวกับ API
func (app *application) About(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "About here!")
}

// ฟังก์ชันสำหรับแสดงรายชื่อหนังทั้งหมดโดยการทดสอบ mock data
func (app *application) AllDemoMovies(w http.ResponseWriter, r *http.Request) {

	// สร้างตัวแปรไว้เก็บข้อมูลหนัง
	var movies []models.Movie

	// กำหนดตัวแปรรูปแบบวันที่ yyyy-mm-dd
	rd, _ := time.Parse("2006-01-02", "1981-06-12")

	// สร้างข้อมูลหนัง
	highlander := models.Movie{
		ID:          1,
		Title:       "Highlander",
		ReleaseDate: rd,
		MPAARating:  "R",
		RunTime:     116,
		Description: "A very nice movie",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// เพิ่มข้อมูลหนังลงใน slice
	movies = append(movies, highlander)

	rd, _ = time.Parse("2006-01-02", "1982-06-07")

	rotla := models.Movie{
		ID:          2,
		Title:       "Raiders of the Lost Ark",
		ReleaseDate: rd,
		MPAARating:  "PG-13",
		RunTime:     115,
		Description: "Another very nice movie",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	movies = append(movies, rotla)

	out, err := json.Marshal(movies)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// AllMovies แสดงรายชื่อหนังทั้งหมด
// @Summary แสดงรายชื่อหนังทั้งหมด
// @Description ดึงข้อมูลหนังทั้งหมดจาก database
// @Tags Movies
// @Produce json
// @Success 200 {array} map[string]interface{} "List of all movies" example([{"id":1,"title":"Movie Title","release_date":"2024-08-28","mpaa_rating":"PG","run_time":120,"description":"Description of the movie"}])
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/movies [get]
func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	// ดึงข้อมูลหนังทั้งหมดจาก database โดยใช้เมธอด AllMovies จาก app.DB
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movies)
}

// GetMovie แสดงรายละเอียดของหนังตาม ID
// @Summary แสดงรายละเอียดของหนังตาม ID
// @Description ดึงข้อมูลหนังตาม ID ที่กำหนด
// @Tags Movies
// @Produce json
// @Param id path int true "Movie ID"
// @Success 200 {object} map[string]interface{} "Movie details" example({"id":1,"title":"Movie Title","release_date":"2024-08-28","mpaa_rating":"PG","run_time":120,"description":"Description of the movie"})
// @Failure 400 {object} map[string]interface{} "Bad Request" example({"error":"Invalid ID"})
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/movies/{id} [get]
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

// MovieForEdit ดึงข้อมูลหนังและประเภทหนังสำหรับการแก้ไข
// @Summary ดึงข้อมูลหนังและประเภทหนังสำหรับการแก้ไข
// @Description ดึงข้อมูลหนังและประเภทหนังสำหรับการแก้ไขตาม ID
// @Tags Movies
// @Produce json
// @Security BearerAuth
// @Param id path int true "Movie ID"
// @Success 200 {object} map[string]interface{} "Movie and genres details" example({"movie":{"id":1,"title":"Movie Title"},"genres":[{"id":1,"name":"Genre Name"}]})
// @Failure 400 {object} map[string]interface{} "Bad Request" example({"error":"Invalid ID"})
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/admin/movies/{id} [get]
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

// MovieCatalog แสดงรายชื่อหนังในแคตตาล็อก
// @Summary แสดงรายชื่อหนังในแคตตาล็อก
// @Description ดึงข้อมูลหนังทั้งหมดจากแคตตาล็อก
// @Tags Movies
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "List of movies in catalog" example([{"id":1,"title":"Catalog Movie Title","release_date":"2024-08-28","mpaa_rating":"PG","run_time":90,"description":"Description of the catalog movie"}])
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/admin/movies [get]
func (app *application) MovieCatalog(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, movies)
}

// AllGenres แสดงประเภทหนังทั้งหมด
// @Summary แสดงประเภทหนังทั้งหมด
// @Description ดึงข้อมูลประเภทหนังทั้งหมด
// @Tags Genres
// @Produce json
// @Success 200 {array} map[string]interface{} "List of all genres" example([{"id":1,"name":"Action"},{"id":2,"name":"Drama"}])
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/genres [get]
func (app *application) AllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.DB.AllGenres()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, genres)
}

// ฟังก์ชันสำหรับดึงรูปภาพหนังจาก API
func (app *application) getPoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
		TotalPages int `json:"total_pages"`
	}

	client := &http.Client{}
	theUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.APIKey)

	// https://api.themoviedb.org/3/search/movie?api_key=b41447e6319d1cd467306735632ba733&query=Die+Hard

	req, err := http.NewRequest("GET", theUrl+"&query="+url.QueryEscape(movie.Title), nil)
	if err != nil {
		log.Println(err)
		return movie
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return movie
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return movie
	}

	var responseObject TheMovieDB

	json.Unmarshal(bodyBytes, &responseObject)

	if len(responseObject.Results) > 0 {
		movie.Image = responseObject.Results[0].PosterPath
	}

	return movie
}

// InsertMovie เพิ่มหนังใหม่
// @Summary เพิ่มหนังใหม่
// @Description เพิ่มหนังใหม่ไปยังฐานข้อมูล
// @Tags Movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body object true "Movie data" example({"title":"New Movie","release_date":"2024-08-28","mpaa_rating":"PG","run_time":120,"description":"New movie description"})
// @Success 202 {object} map[string]interface{} "Movie created" example({"message":"movie updated"})
// @Failure 400 {object} map[string]interface{} "Bad Request" example({"error":"Invalid data"})
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/admin/movies [post]
func (app *application) InsertMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie

	err := app.readJSON(w, r, &movie)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// try to get an image
	movie = app.getPoster(movie)

	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	newID, err := app.DB.InsertMovie(movie)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// now handle genres
	err = app.DB.UpdateMovieGenres(newID, movie.GenresArray)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}

// UpdateMovie แก้ไขข้อมูลหนัง
// @Summary แก้ไขข้อมูลหนัง
// @Description แก้ไขข้อมูลหนังตาม ID ที่กำหนด
// @Tags Movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body object true "Updated movie data" example({"id":1,"title":"Updated Movie Title","release_date":"2024-08-28","mpaa_rating":"PG","run_time":130,"description":"Updated movie description"})
// @Success 202 {object} map[string]interface{} "Movie updated" example({"message":"movie updated"})
// @Failure 400 {object} map[string]interface{} "Bad Request" example({"error":"Invalid data"})
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/admin/movies/{id} [put]
func (app *application) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	var payload models.Movie

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movie, err := app.DB.OneMovie(payload.ID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movie.Title = payload.Title
	movie.ReleaseDate = payload.ReleaseDate
	movie.Description = payload.Description
	movie.MPAARating = payload.MPAARating
	movie.RunTime = payload.RunTime
	movie.UpdatedAt = time.Now()

	err = app.DB.UpdateMovie(*movie)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.DB.UpdateMovieGenres(movie.ID, payload.GenresArray)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}

// DeleteMovie ลบหนังตาม ID
// @Summary ลบหนังตาม ID
// @Description ลบข้อมูลหนังตาม ID ที่กำหนด
// @Tags Movies
// @Produce json
// @Security BearerAuth
// @Param id path int true "Movie ID"
// @Success 202 {object} map[string]interface{} "Movie deleted" example({"message":"movie deleted"})
// @Failure 400 {object} map[string]interface{} "Bad Request" example({"error":"Invalid ID"})
// @Failure 500 {object} map[string]interface{} "Internal Server Error" example({"error":"Internal Server Error"})
// @Router /api/v1/admin/movies/{id} [delete]
func (app *application) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.DB.DeleteMovie(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie deleted",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}
