package imaged

import (
	"context"
	"github.com/jmoiron/sqlx"
	// Pull in the PostgreSQL driver
	_ "github.com/lib/pq"
)

// DBConn is a handle to a connection to the database.
type DBConn struct {
	*sqlx.DB
}

// NewDBConn creates a new database connection handle.
func NewDBConn(url string) (*DBConn, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &DBConn{
		db,
	}, nil
}

// RecentBuilds gets a list of the most recent builds.
func (db *DBConn) RecentBuilds(ctx context.Context) ([]Build, error) {
	var builds []Build
	if err := db.Select(&builds, "SELECT * FROM builds ORDER BY id DESC LIMIT 20"); err != nil {
		return nil, err
	}

	return builds, nil
}

// CreateBuild records a new build that was just requested.
func (db *DBConn) CreateBuild(ctx context.Context, name string, revision string) (*Build, error) {
	var id int64
	err := db.QueryRowContext(ctx, "INSERT INTO builds (name, revision) VALUES ($1, $2) RETURNING id", name, revision).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &Build{
		ID:       id,
		Name:     name,
		Revision: revision,
	}, err
}
