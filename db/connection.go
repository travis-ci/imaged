package db

import (
	"github.com/jmoiron/sqlx"
	// Pull in the PostgreSQL driver
	_ "github.com/lib/pq"
)

// Connection is a handle to a connection to the database.
type Connection struct {
	*sqlx.DB
}

// NewConnection creates a new database connection handle.
func NewConnection(url string) (*Connection, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &Connection{
		db,
	}, nil
}
