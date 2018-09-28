package imaged

import (
	"context"
	pb "github.com/travis-ci/imaged/rpc/images"
	"github.com/twitchtv/twirp"
)

// Server handles API requests for imaged.
type Server struct {
	DB      *DBConn
	Storage *Storage
}

// NewServer creates a new handler for API requests.
func NewServer(databaseURL string, recordBucket string) (*Server, error) {
	db, err := NewDBConn(databaseURL)
	if err != nil {
		return nil, err
	}

	storage, err := NewStorage(recordBucket)
	if err != nil {
		return nil, err
	}

	return &Server{
		DB:      db,
		Storage: storage,
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

// DownloadRecord downloads the file contents of a build record from S3.
func (s *Server) DownloadRecord(ctx context.Context, req *pb.DownloadRecordRequest) (*pb.DownloadRecordResponse, error) {
	r, err := s.DB.GetRecord(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	b, err := s.Storage.DownloadBytes(ctx, r.S3Key)
	if err != nil {
		return nil, err
	}

	resp := &pb.DownloadRecordResponse{
		Contents: b,
	}

	return resp, nil
}

// AttachRecord creates a new build record based on a file that was generated during the build.
//
// Generally, this won't need to be used, as imaged should upload records itself after a build runs. This API method is for backfilling records.
func (s *Server) AttachRecord(ctx context.Context, req *pb.AttachRecordRequest) (*pb.AttachRecordResponse, error) {
	if req.FileName == "" {
		return nil, twirp.InvalidArgumentError("file_name", "cannot be empty")
	}

	build, err := s.DB.GetBuild(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	key := build.RecordKey(req.FileName)
	if _, err = s.Storage.UploadBytes(ctx, key, req.Contents); err != nil {
		return nil, err
	}

	record, err := s.DB.CreateRecord(ctx, build, req.FileName, key)
	if err != nil {
		return nil, err
	}

	resp := &pb.AttachRecordResponse{
		Record: record.Message(),
	}

	return resp, nil
}
