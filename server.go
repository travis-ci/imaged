package imaged

import (
	"context"
	pb "github.com/travis-ci/imaged/rpc/images"
)

// Server handles API requests for imaged.
type Server struct {
	DB *DBConn
}

// NewServer creates a new handler for API requests.
func NewServer(databaseURL string) (*Server, error) {
	db, err := NewDBConn(databaseURL)
	if err != nil {
		return nil, err
	}

	return &Server{
		DB: db,
	}, nil
}

// ListBuilds provides a list of recent builds that imaged has run.
func (s *Server) ListBuilds(ctx context.Context, req *pb.ListBuildsRequest) (*pb.ListBuildsResponse, error) {
	builds, err := s.DB.RecentBuilds(ctx)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListBuildsResponse{}
	for _, b := range builds {
		resp.Builds = append(resp.Builds, b.Message())
	}

	return resp, nil
}

// StartBuild creates a new build and begins running it.
func (s *Server) StartBuild(ctx context.Context, req *pb.StartBuildRequest) (*pb.StartBuildResponse, error) {
	build, err := s.DB.CreateBuild(ctx, req.Name, req.Revision)
	if err != nil {
		return nil, err
	}

	resp := &pb.StartBuildResponse{
		Build: build.Message(),
	}

	return resp, nil
}
