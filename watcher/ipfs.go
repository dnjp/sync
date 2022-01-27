package watcher

import (
	"bytes"
	"context"
	"net/http"

	shell "github.com/ipfs/go-ipfs-api"
)

// NewIPFSClient constructs an http client configured to perform basic
// auth on every request
func NewIPFSClient(projectID, projectSecret string) *http.Client {
	return &http.Client{
		Transport: authTransport{
			RoundTripper:  http.DefaultTransport,
			ProjectId:     projectID,
			ProjectSecret: projectSecret,
		},
	}
}

type authTransport struct {
	http.RoundTripper
	ProjectId     string
	ProjectSecret string
}

// RoundTrip satisfies the http.RoundTripper interface, allowing the transport
// to configure auth on every request
func (t authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.ProjectId, t.ProjectSecret)
	return t.RoundTripper.RoundTrip(r)
}

func cat(ctx context.Context, sh *shell.Shell, path string) ([]byte, error) {
	recvData := make(chan []byte)
	errs := make(chan error)

	go func(recvData chan []byte, errs chan error) {
		var data bytes.Buffer
		closer, err := sh.Cat(path)
		if err != nil {
			errs <- err
			return
		}
		data.ReadFrom(closer)
		if err := closer.Close(); err != nil {
			errs <- err
			return
		}
		recvData <- data.Bytes()
	}(recvData, errs)

	select {
	case data := <-recvData:
		return data, nil
	case err := <-errs:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
