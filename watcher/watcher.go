package watcher

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"
)

// Watcher is a utility that can watch for updates
type Watcher struct {
	file *File
	stop chan struct{}
}

// NewWatcher constructs a new watcher for the given file
func NewWatcher(file *File) *Watcher {
	w := &Watcher{
		file: file,
		stop: make(chan struct{}),
	}
	file.attachWatcher(w)
	return w
}

// Stop stops the watcher from watching the file
func (w *Watcher) Stop() {
	w.stop <- struct{}{}
	w.file.detachWatcher()
}

// Start causes the watcher to begin watching the file for changes
func (w *Watcher) Start(
	interval time.Duration,
	errs chan<- error,
	removals chan<- FilePath,
	additions chan<- FilePath,
) {
	go w.watch(interval, errs, removals, additions)
}

func (w *Watcher) watch(
	interval time.Duration,
	errs chan<- error,
	removals chan<- FilePath,
	additions chan<- FilePath,
) {

	checksum, err := w.file.Checksum()
	if err != nil {
		log.Printf("ERR: %+v\n", err)
		errs <- err
		return
	}

	tick := time.NewTicker(interval)
	internalErrs := make(chan error)
	checksumUpdates := make(chan [32]byte)
	var mux sync.RWMutex

	for {
		select {
		case <-w.stop:
			return
		case err := <-internalErrs:
			log.Printf("ERR: %+v\n", err)
			errs <- err
			return
		case updatedChecksum := <-checksumUpdates:
			log.Printf("file %s changed\n", w.file.AbsolutePath)
			mux.Lock()
			checksum = updatedChecksum
			mux.Unlock()
		case <-tick.C:
			_, err := os.Stat(w.file.AbsolutePath.String())
			if errors.Is(err, os.ErrNotExist) {
				log.Printf("file %s has been removed\n", w.file.AbsolutePath)
				removals <- w.file.AbsolutePath
				return
			} else if err != nil {
				errs <- err
				return
			}
			go w.checkFileUpdated(checksum, internalErrs, checksumUpdates, additions)
		}
	}
}

func (w *Watcher) checkFileUpdated(
	checksum [32]byte,
	errs chan<- error,
	updates chan<- [32]byte,
	additions chan<- FilePath,
) {
	cs, err := w.file.Checksum()
	if err != nil {
		errs <- err
		return
	}

	if cs != checksum {
		additions <- w.file.AbsolutePath
		updates <- cs
	}
}
