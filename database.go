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

// GetBuild retrieves a build by ID.
func (db *DBConn) GetBuild(ctx context.Context, id int64) (*Build, error) {
	var build Build
	if err := db.Get(&build, "SELECT * FROM builds WHERE id = $1", id); err != nil {
		return nil, err
	}

	return &build, nil
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

// GetRecord retrieves a build record by ID.
func (db *DBConn) GetRecord(ctx context.Context, id int64) (*Record, error) {
	var record Record
	if err := db.Get(&record, "SELECT * FROM records WHERE id = $1", id); err != nil {
		return nil, err
	}

	return &record, nil
}

// CreateRecord records a new build record that has already been uploaded to S3.
func (db *DBConn) CreateRecord(ctx context.Context, build *Build, filename string, s3key string) (*Record, error) {
	var id int64
	if err := db.QueryRowContext(ctx, "INSERT INTO records (build_id, filename, s3_key) VALUES ($1, $2, $3) RETURNING id", build.ID, filename, s3key).Scan(&id); err != nil {
		return nil, err
	}

	return &Record{
		ID:       id,
		BuildID:  build.ID,
		FileName: filename,
		S3Key:    s3key,
	}, nil
}
