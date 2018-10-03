package server

import (
	"context"
	"github.com/travis-ci/imaged/db"
	pb "github.com/travis-ci/imaged/rpc/images"
	"github.com/travis-ci/imaged/storage"
	"github.com/travis-ci/imaged/worker"
	"github.com/twitchtv/twirp"
)

// Server handles API requests for imaged.
type Server struct {
	DB      *db.Connection
	Storage *storage.Storage
	Worker  *worker.Worker
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

	s.Worker.Send(worker.Job{Build: build})

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

// GetRecordURL creates a temporary public URL for downloading the build record.
func (s *Server) GetRecordURL(ctx context.Context, req *pb.GetRecordURLRequest) (*pb.GetRecordURLResponse, error) {
	r, err := s.DB.GetRecord(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	url, err := s.Storage.PublicURL(ctx, r.S3Key)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetRecordURLResponse{
		Url: url,
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
