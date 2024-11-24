package repository

import (
	"backend/internal/models"
	"database/sql"
)

type DatabaseRepo interface {
	Conconnection() *sql.DB
	//Get all movie

	AllMovies() ([]*models.Movie, error)
}
