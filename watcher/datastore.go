package watcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	wraperr "github.com/pkg/errors"
)

type store map[FilePath]*File

// Datastore wraps a distributed datastore like IPFS, but keeps
// all watched files up to date
type Datastore struct {
	// handles
	sh    *shell.Shell
	store store
	// communication
	errs      chan error
	additions chan FilePath
	removals  chan FilePath
	stop      chan struct{}
	// state
	cid CID
	// settings
	backupLocation      string
	fileRefreshInterval time.Duration
	// synchronization
	mux sync.RWMutex
}

// NewDatastore constructs a datastore with the given settings
func NewDatastore(sh *shell.Shell, backupLocation string, refreshInterval time.Duration) (*Datastore, error) {
	datastore := &Datastore{
		// handles
		sh:    sh,
		store: make(store),
		// communication
		errs:      make(chan error),
		stop:      make(chan struct{}),
		additions: make(chan FilePath),
		removals:  make(chan FilePath),
		// settings
		backupLocation:      backupLocation,
		fileRefreshInterval: refreshInterval,
	}

	cidBytes, err := ioutil.ReadFile(backupLocation)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		data, err := cat(ctx, sh, string(cidBytes))
		if err != nil {
			log.Printf("could not retrieve previous cid: %+v\n", err)
			log.Println("creating store from scratch")
			return datastore, nil
		}

		err = datastore.FromJSON(data)
		if err != nil {
			return nil, err
		}
	}

	return datastore, nil
}

// Start launches the event listeners for the datastore, returning
// the first error returned
func (d *Datastore) Start() error {
	go d.listenAdditions(d.additions)
	go d.listenRemovals(d.removals)
	select {
	case err := <-d.errs:
		return err
	case <-d.stop:
		return nil
	}
}

// Stop gracefully stops all event processing within the datastore
func (d *Datastore) Stop() error {
	d.mux.RLock()
	for _, file := range d.store {
		if file.Watcher != nil {
			file.Watcher.Stop()
		}
	}
	d.mux.RUnlock()
	d.stop <- struct{}{}
	return nil
}

func (d *Datastore) listenAdditions(newFiles chan FilePath) {
	for {
		path := <-newFiles
		_, err := d.AddFile(path)
		if err != nil {
			d.errs <- err
		}
	}
}

func (d *Datastore) listenRemovals(removedFiles chan FilePath) {
	for {
		err := d.RemoveFile(<-removedFiles)
		if err != nil {
			d.errs <- err
		}
	}
}

func (d *Datastore) RemoveFile(path FilePath) error {
	if path != "" {
		delete(d.store, path)
		if err := d.commit(); err != nil {
			return err
		}
	}
	return nil
}

// AddFile adds the file at the given path to the datastore
func (d *Datastore) AddFile(path FilePath) (*File, error) {

	var file *File

	d.mux.RLock()
	f, fileExists := d.store[path]
	d.mux.RUnlock()

	if fileExists && f != nil {
		file = f
	} else {
		f, err := NewFile(path)
		if err != nil {
			return nil, err
		}
		file = f
	}

	b, err := file.Read()
	if err != nil {
		return nil, err
	}

	cid, err := d.sh.Add(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	file.AssignCID(CID(cid))
	err = d.sh.Pin(cid)
	if err != nil {
		return file, err
	}

	d.mux.Lock()
	d.store[file.AbsolutePath] = file
	d.mux.Unlock()

	if err := d.commit(); err != nil {
		return file, err
	}

	if !fileExists {
		NewWatcher(file).Start(
			d.fileRefreshInterval,
			d.errs,
			d.removals,
			d.additions,
		)
	}

	return file, nil
}

// Add adds the file or directory at the given path to the store while communicating
// any errors that are encountered. When all files have been processed a message
// is published one the "done" channel
func (d *Datastore) Add(path FilePath) (files chan *File, done chan struct{}, errs chan error) {

	done = make(chan struct{})
	files = make(chan *File)
	errs = make(chan error)

	go func(path FilePath, files chan *File, errs chan error) {

		info, err := os.Stat(path.String())
		if err != nil {
			errs <- err
			return
		}

		if info.IsDir() {
			// recursively add all files within the directory
			err := filepath.Walk(path.String(), func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				file, err := d.AddFile(FilePath(path))
				if err != nil {
					return err
				}
				files <- file
				return nil
			})
			if err != nil {
				errs <- err
				return
			}
			done <- struct{}{}
		} else {
			// add the single file
			file, err := d.AddFile(FilePath(path))
			if err != nil {
				errs <- err
				return
			}
			files <- file
			done <- struct{}{}
		}
	}(path, files, errs)

	return
}

func (d *Datastore) commit() error {

	b, err := d.JSON()
	if err != nil {
		return err
	}

	cid, err := d.sh.Add(bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	var errs []error
	if err := d.sh.Pin(cid); err != nil {
		errs = append(errs, err)
	}

	d.UpdateCID(CID(cid))
	err = d.backupCID()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		err := fmt.Errorf("errors encountered while commiting changes: %w;", errs[0])
		for i := 1; i < len(errs); i++ {
			err = wraperr.Wrap(err, errs[i].Error())
		}
		return err
	}

	return nil
}

func (d *Datastore) backupCID() error {
	cid, ok := d.CID()
	if !ok {
		return fmt.Errorf("cid is not set")
	}
	return os.WriteFile(d.backupLocation, []byte(cid), 0644)
}

// CID returns the content indentifier for all store metadata
func (d *Datastore) CID() (CID, bool) {
	d.mux.RLock()
	defer d.mux.RUnlock()
	if d.cid == "" {
		return "", false
	}
	return d.cid, true
}

// UpdateCID updates the content identifier for the datastore
func (d *Datastore) UpdateCID(cid CID) {
	d.mux.Lock()
	d.cid = cid
	d.mux.Unlock()
}

// FindCID returns the CID matching the given path if found
func (d *Datastore) FindCID(path FilePath) (CID, bool) {
	d.mux.RLock()
	defer d.mux.RUnlock()
	file, ok := d.store[path]
	if !ok {
		return "", false
	}
	return file.CID, ok
}

// FindPath returns the file path matching the given CID if found
func (d *Datastore) FindPath(cid CID) (FilePath, bool) {
	d.mux.RLock()
	defer d.mux.RUnlock()
	given := cid
	for path, file := range d.store {
		if file.CID == given {
			return path, true
		}
	}
	return "", false
}

// JSON returns the JSON representation of the files within the store
func (d *Datastore) JSON() ([]byte, error) {
	d.mux.RLock()
	defer d.mux.RUnlock()
	return json.Marshal(d.store)
}

// FromJSON populates the datastore with files from a JSON payload
func (d *Datastore) FromJSON(b []byte) error {

	tmp := make(store)
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

Restore:
	for path, restoreFile := range tmp {

		d.mux.RLock()
		file, ok := d.store[path]
		d.mux.RUnlock()

		if ok && file.CID != restoreFile.CID {
			return fmt.Errorf(
				"conflict for file %s. current cid is %s, cid from IPFS is %s",
				path,
				file.CID,
				restoreFile.CID,
			)
		}

		files, done, errs := d.Add(path)
		for {
			select {
			case <-files:
			case <-done:
				continue Restore
			case err := <-errs:
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// RangeStore iterates over all files in the store until done() returns true
func (d *Datastore) RangeStore(done func(*File) (done bool)) {
	d.mux.RLock()
	defer d.mux.RUnlock()
	for _, file := range d.store {
		if done(file) {
			return
		}
	}
}
