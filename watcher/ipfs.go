package watcher

import (
	"bytes"
	"context"

	shell "github.com/ipfs/go-ipfs-api"
)

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
