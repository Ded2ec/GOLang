package dbrepo

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"time"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 5

// Conconnection

func (m *PostgresDBRepo) Conconnection() *sql.DB {
	return m.DB
}

// AllMovie return all movie from database
func (m *PostgresDBRepo) AllMovies() ([]*models.Movie, error) {
	// context to handle db query timeout

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	//query db postgres

	query := `
		select
			id, title, release_date, runtime,
			mpaa_rating, description, coalesce(image, ''),
			created_at, updated_at
		from
			movies
		order by
			title
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// declare a slice to hold the models.Movie objects
	var movies []*models.Movie

	// loop through the rows and scan the data into the models.Movie object
	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.RunTime,
			&movie.MPAARating,
			&movie.Description,
			&movie.Image,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}
