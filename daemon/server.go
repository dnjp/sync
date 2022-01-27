package daemon

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/dnjp/zync/proto/zync/v1"
	"github.com/dnjp/zync/watcher"
	shell "github.com/ipfs/go-ipfs-api"
	"google.golang.org/grpc"
)

// Server provides a gRPC interface for interacting
// with watched files stored in IPFS
type Server struct {
	srv   *grpc.Server
	lis   net.Listener
	store *watcher.Datastore
	zync.UnimplementedZyncServer
}

// NewServer constructs a new gPRC server for the daemon
func NewServer(port int, sh *shell.Shell, backupLoc string, refreshInterval time.Duration) (*Server, error) {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	store, err := watcher.NewDatastore(sh, backupLoc, refreshInterval)
	if err != nil {
		return nil, err
	}

	s := &Server{
		srv:   grpc.NewServer(),
		lis:   lis,
		store: store,
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
	errs := make(chan error)
	go func() { errs <- s.store.Start() }()
	go func() { errs <- s.srv.Serve(s.lis) }()
	return <-errs
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() error {
	s.srv.GracefulStop()
	log.Println("grpc server stopped")
	return s.store.Stop()
}

func isFilePath(path string) bool {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// AddFiles adds a single file to IPFS if the provided
// path is to an individual file, or it will recursively
// add all files within a directory if the provided path
// is a directory
func (s *Server) AddFiles(req *zync.RegexRequest, afs zync.Zync_AddFilesServer) error {

	if req.Pattern == "" {
		return fmt.Errorf("must provide pattern")
	}

	// received an absolute path
	if isFilePath(req.Pattern) {
		file, err := s.store.AddFile(watcher.FilePath(req.Pattern))
		if err != nil {
			return err
		}
		if err := afs.Send(file.Status()); err != nil {
			return err
		}
	} else {
		regex, err := regexp.Compile(req.Pattern)
		if err != nil {
			return err
		}

		err = filepath.Walk(req.CurrentDirectory, func(path string, info fs.FileInfo, err error) error {
			if regex.MatchString(path) {
				file, err := s.store.AddFile(watcher.FilePath(path))
				if err != nil {
					return err
				}
				if err := afs.Send(file.Status()); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ListFiles lists all files matching the pattern from
// zync
func (s *Server) ListFiles(req *zync.RegexRequest, lfs zync.Zync_ListFilesServer) error {
	var regex *regexp.Regexp
	if req.Pattern != "" {
		var err error
		regex, err = regexp.Compile(req.Pattern)
		if err != nil {
			return err
		}
	}

	var returnErr error
	s.store.RangeStore(func(file *watcher.File) (done bool) {
		shouldSend := false
		if regex != nil {
			if regex.MatchString(file.AbsolutePath.String()) {
				shouldSend = true
			}
		} else {
			shouldSend = true
		}
		if shouldSend {
			if err := lfs.Send(file.Status()); err != nil {
				returnErr = err
				return true
			}
		}
		return false
	})
	if returnErr != nil {
		return returnErr
	}

	return nil
}

// DeleteFiles removes all files matching the pattern
// from zync
func (s *Server) DeleteFiles(req *zync.RegexRequest, dfs zync.Zync_DeleteFilesServer) error {
	var regex *regexp.Regexp
	if req.Pattern != "" {
		var err error
		regex, err = regexp.Compile(req.Pattern)
		if err != nil {
			return err
		}
	}

	filesToRemove := make(map[watcher.FilePath]*watcher.File)

	var returnErr error
	s.store.RangeStore(func(file *watcher.File) (done bool) {
		shouldSend := false
		if regex != nil {
			if regex.MatchString(file.AbsolutePath.String()) {
				shouldSend = true
			}
		} else {
			shouldSend = true
		}
		if shouldSend {
			filesToRemove[file.AbsolutePath] = file
		}
		return false
	})
	if returnErr != nil {
		return returnErr
	}

	for path, file := range filesToRemove {
		log.Printf("removing file %s\n", path)
		if err := s.store.RemoveFile(path); err != nil {
			return err
		}
		if err := dfs.Send(file.Status()); err != nil {
			return err
		}
	}

	return nil
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
