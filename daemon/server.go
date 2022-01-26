package daemon

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/dnjp/zync/proto/zync/v1"
	"google.golang.org/grpc"
)

// Server provides a gRPC interface for interacting
// with watched files stored in IPFS
type Server struct {
	srv *grpc.Server
	lis net.Listener
	zync.UnimplementedZyncServer
}

// NewServer constructs a new gPRC server for the daemon
func NewServer(port int) (*Server, error) {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		srv: grpc.NewServer(),
		lis: lis,
	}
	zync.RegisterZyncServer(s.srv, s)

	return s, nil
}

// HandleTerminate stops the server if terminate signal
// is received
func (s *Server) HandleTerminate(sig os.Signal) error {
	switch sig {
	case syscall.SIGQUIT:
	case syscall.SIGTERM:
		return s.Stop()
	}
	return nil
}

// Start launches the gRPC server
func (s *Server) Start() error {
	log.Println("grpc server listening...")
	return s.srv.Serve(s.lis)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() error {
	s.srv.GracefulStop()
	log.Println("grpc server stopped")
	return nil
}

// AddFiles adds a single file to IPFS if the provided
// path is to an individual file, or it will recursively
// add all files within a directory if the provided path
// is a directory
func (s *Server) AddFiles(req *zync.RegexRequest, afs zync.Zync_AddFilesServer) error {
	return fmt.Errorf("TODO")
}

// ListFiles lists all files matching the pattern from
// zync
func (s *Server) ListFiles(req *zync.RegexRequest, lfs zync.Zync_ListFilesServer) error {
	return fmt.Errorf("TODO")
}

// DeleteFiles removes all files matching the pattern
// from zync
func (s *Server) DeleteFiles(req *zync.RegexRequest, dfs zync.Zync_DeleteFilesServer) error {
	return fmt.Errorf("TODO")
}

// Backup communicates to the server that any cached
// data should be backed up to IPFS, returning the
// resulting CID
func (s *Server) Backup(ctx context.Context, req *zync.BackupRequest) (*zync.BackupStatus, error) {
	return nil, fmt.Errorf("TODO")
}

// Restore initiates the process of restoring files
// from IPFS to the host machine
func (s *Server) Restore(req *zync.RestoreRequest, rs zync.Zync_RestoreServer) error {
	return fmt.Errorf("TODO")
}
