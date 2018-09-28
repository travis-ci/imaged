package imaged

import (
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
