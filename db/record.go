package db

import (
	"context"
	pb "github.com/travis-ci/imaged/rpc/images"
)

// Record represents a build artifact or other file produced during a
// Packer build that should be kept for reference later.
type Record struct {
	ID       int64
	BuildID  int64 `db:"build_id"`
	FileName string
	S3Key    string `db:"s3_key"`
}

// Message converts the record into a protobuf message.
func (r *Record) Message() *pb.Record {
	return &pb.Record{
		Id:       r.ID,
		BuildId:  r.BuildID,
		FileName: r.FileName,
		S3Key:    r.S3Key,
	}
}

// GetRecord retrieves a build record by ID.
func (db *Connection) GetRecord(ctx context.Context, id int64) (*Record, error) {
	var record Record
	if err := db.Get(&record, "SELECT * FROM records WHERE id = $1", id); err != nil {
		return nil, err
	}

	return &record, nil
}

// CreateRecord records a new build record that has already been uploaded to S3.
func (db *Connection) CreateRecord(ctx context.Context, build *Build, filename string, s3key string) (*Record, error) {
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
