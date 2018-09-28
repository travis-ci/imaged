package imaged

import (
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
