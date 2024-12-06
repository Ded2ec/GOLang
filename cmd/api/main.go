package main

import (
	"backend/internal/repository"
	"backend/internal/repository/dbrepo"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	// "github.com/joho/godotenv"
)

// Set Port
const port = 8080

// Appilication Struct&config Database connection

type application struct {
	DSN          string
	Domain       string
	DB           repository.DatabaseRepo
	auth         Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
	APIKey       string
}

// @title Movies API with GO and PostgreSQL
// @version 1.0
// @description This is a Movies API with GO and PostgreSQL
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	// Set appilication Config
	var app application
	app.Domain = "example.com"

	// Read from command line agements
	//flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 dbname=gosampledb user=postgres password=15229 sslmode=disable timezone=UTC connect_timeout=5", "Postgres connect db")

	// Load .env

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// สร้าง DSN สำหรับเชื่อมต่อฐานข้อมูล
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s timezone=%s connect_timeout=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
		os.Getenv("DB_CONNECT_TIMEOUT"),
	)
	// อ่านค่าจาก environment variables
	app.DSN = dsn

	// ตั้งค่าตัวแปร JWT จาก environment variables
	app.JWTSecret = os.Getenv("JWT_SECRET")
	app.JWTIssuer = os.Getenv("JWT_ISSUER")
	app.JWTAudience = os.Getenv("JWT_AUDIENCE")
	app.CookieDomain = os.Getenv("COOKIE_DOMAIN")
	app.Domain = os.Getenv("DOMAIN")
	app.APIKey = os.Getenv("API_KEY")
	// Parse the command line arguments for JWT
	// flag.StringVar(&app.JWTSecret, "jwt-secret", "newty_secret", "signing secret")
	// flag.StringVar(&app.JWTIssuer, "jwt-issuer", "example.com", "signing issuer")
	// flag.StringVar(&app.JWTAudience, "jwt-audience", "example.com", "signing audience")
	// flag.StringVar(&app.CookieDomain, "cookie-domain", "localhost", "cookie domain")
	// flag.StringVar(&app.Domain, "domain", "example.com", "domain")
	// flag.StringVar(&app.APIKey, "api-key", "b41447e6319d1cd467306735632ba733", "api key")

	flag.Parse()

	// Connect to Database
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	defer app.DB.Conconnection().Close()

	app.auth = Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		CookieName:    "__Host-refresh_token",
		CookieDomain:  app.CookieDomain,
	}

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
