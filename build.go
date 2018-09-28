package imaged

import (
	pb "github.com/travis-ci/imaged/rpc/images"
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
