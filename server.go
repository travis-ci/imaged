package imaged

import (
	"context"
	pb "github.com/travis-ci/imaged/rpc/images"
)

// Server handles API requests for imaged.
type Server struct {
}

// ListBuilds provides a list of recent builds that imaged has run.
func (s *Server) ListBuilds(ctx context.Context, req *pb.ListBuildsRequest) (*pb.ListBuildsResponse, error) {
	return &pb.ListBuildsResponse{}, nil
}
