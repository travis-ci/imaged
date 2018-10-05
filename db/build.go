package db

import (
	"context"
	"database/sql/driver"
	"github.com/pkg/errors"
	pb "github.com/travis-ci/imaged/rpc/images"
	"strconv"
	"strings"
	"time"
)

// BuildStatus is an enumeration of possible statuses a build could have.
type BuildStatus string

// These are the defined statuses a build could have.
//
// Since these are represented with a PG enum type, we should never get a
// different value from the DB.
const (
	BuildStatusCreated   BuildStatus = "created"
	BuildStatusStarted   BuildStatus = "started"
	BuildStatusSucceeded BuildStatus = "succeeded"
	BuildStatusFailed    BuildStatus = "failed"
)

// Build represents a Packer build that a user requested to run.
type Build struct {
	ID           int64
	Name         string
	Revision     string
	FullRevision *string `db:"full_revision"`
	Status       BuildStatus
	CreatedAt    time.Time  `db:"created_at"`
	StartedAt    *time.Time `db:"started_at"`
	FinishedAt   *time.Time `db:"finished_at"`
	Records      []Record
}

// Message converts the build into a protobuf message.
func (b *Build) Message() *pb.Build {
	var fullRevision string
	if b.FullRevision != nil {
		fullRevision = *b.FullRevision
	}

	var start, finish int64
	if b.StartedAt != nil {
		start = b.StartedAt.Unix()
	}
	if b.FinishedAt != nil {
		finish = b.FinishedAt.Unix()
	}

	msg := &pb.Build{
		Id:           b.ID,
		Name:         b.Name,
		Revision:     b.Revision,
		FullRevision: fullRevision,
		Status:       b.Status.Enum(),
		CreatedAt:    b.CreatedAt.Unix(),
		StartedAt:    start,
		FinishedAt:   finish,
	}
	for _, r := range b.Records {
		msg.Records = append(msg.Records, r.Message())
	}
	return msg
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

// GetBuildFull retreives a build by ID, and its attached records.
func (db *Connection) GetBuildFull(ctx context.Context, id int64) (*Build, error) {
	build, err := db.GetBuild(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = db.Select(&build.Records, "SELECT * FROM records WHERE build_id = $1", id); err != nil {
		return nil, err
	}

	return build, nil
}

// CreateBuild records a new build that was just requested.
func (db *Connection) CreateBuild(ctx context.Context, name string, revision string) (*Build, error) {
	var id int64
	err := db.QueryRowContext(ctx, "INSERT INTO builds (name, revision) VALUES ($1, $2) RETURNING id", name, revision).Scan(&id)
	if err != nil {
		return nil, err
	}

	return db.GetBuild(ctx, id)
}

// StartBuild marks a build as started and updates its started at timestamp.
func (db *Connection) StartBuild(ctx context.Context, b *Build) error {
	if _, err := db.ExecContext(ctx, "UPDATE builds SET status = 'started', started_at = now() WHERE id = $1", b.ID); err != nil {
		return err
	}

	newBuild, err := db.GetBuild(ctx, b.ID)
	if err != nil {
		return err
	}

	*b = *newBuild
	return nil
}

// FinishBuild marks a build as passed or failed and updates its finished at timestamp.
func (db *Connection) FinishBuild(ctx context.Context, b *Build) error {
	switch b.Status {
	default:
		return errors.New("build must be either succeeded or failed to be finished")
	case BuildStatusSucceeded, BuildStatusFailed:
		if _, err := db.ExecContext(ctx, "UPDATE builds SET status = $2, finished_at = now() WHERE id = $1", b.ID, b.Status); err != nil {
			return err
		}

		newBuild, err := db.GetBuild(ctx, b.ID)
		if err != nil {
			return err
		}

		*b = *newBuild
		return nil
	}
}

// UpdateBuild updates some fields about a build.
//
// Currently, this only updates the full revision.
func (db *Connection) UpdateBuild(ctx context.Context, b *Build) error {
	if _, err := db.ExecContext(ctx, "UPDATE builds SET full_revision = $2 WHERE id = $1", b.ID, b.FullRevision); err != nil {
		return err
	}

	return nil
}

// Scan reads a build status from a database type.
func (s *BuildStatus) Scan(value interface{}) error {
	bytes := value.([]byte)
	*s = BuildStatus(string(bytes))
	return nil
}

// Value converts the status to a string for the sql package.
func (s BuildStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Enum converts the status to a protobuf enum.
func (s BuildStatus) Enum() pb.Build_Status {
	str := strings.ToUpper(string(s))
	val := pb.Build_Status_value[str]
	return pb.Build_Status(val)
}
