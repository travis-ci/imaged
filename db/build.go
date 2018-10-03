package db

import (
	"context"
	pb "github.com/travis-ci/imaged/rpc/images"
	"strconv"
)

// Build represents a Packer build that a user requested to run.
type Build struct {
	ID       int64
	Name     string
	Revision string
}

// Message converts the build into a protobuf message.
func (b *Build) Message() *pb.Build {
	return &pb.Build{
		Id:       b.ID,
		Name:     b.Name,
		Revision: b.Revision,
	}
}

// RecordKey generates an S3 key for storing a build record for this build.
func (b *Build) RecordKey(filename string) string {
	return "records/" + strconv.FormatInt(b.ID, 10) + "/" + filename
}

// RecentBuilds gets a list of the most recent builds.
func (db *Connection) RecentBuilds(ctx context.Context) ([]Build, error) {
	var builds []Build
	if err := db.Select(&builds, "SELECT * FROM builds ORDER BY id DESC LIMIT 20"); err != nil {
		return nil, err
	}

	return builds, nil
}

// GetBuild retrieves a build by ID.
func (db *Connection) GetBuild(ctx context.Context, id int64) (*Build, error) {
	var build Build
	if err := db.Get(&build, "SELECT * FROM builds WHERE id = $1", id); err != nil {
		return nil, err
	}

	return &build, nil
}

// CreateBuild records a new build that was just requested.
func (db *Connection) CreateBuild(ctx context.Context, name string, revision string) (*Build, error) {
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
