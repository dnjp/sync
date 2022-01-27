package watcher

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"sync"

	"github.com/dnjp/zync/proto/zync/v1"
)

// CID represents an IPFS content identifier. See https://docs.ipfs.io/concepts/content-addressing/
type CID string

func (cid CID) String() string {
	return string(cid)
}

// FilePath represents a proper file path
type FilePath string

func (path FilePath) String() string {
	return string(path)
}

// File represents a file being watched
type File struct {
	CID          CID      `json:"cid"`
	AbsolutePath FilePath `json:"absolute_path"`
	Watcher      *Watcher `json:"-"`
	checksum     [32]byte
	data         *bytes.Buffer
	mux          sync.RWMutex
}

// NewFile constructs a File, updating its properties from disk
func NewFile(path FilePath) (*File, error) {
	f := &File{
		AbsolutePath: path,
		data:         new(bytes.Buffer),
	}
	_, err := f.Read()
	if err != nil {
		return nil, err
	}
	return f, nil
}

// AssignCID updates the CID reference in the File
func (f *File) AssignCID(cid CID) {
	f.mux.Lock()
	f.CID = cid
	f.mux.Unlock()
}

// Read reads the contents of the file, updating its current checksum
func (f *File) Read() ([]byte, error) {

	b, err := ioutil.ReadFile(f.AbsolutePath.String())
	if err != nil {
		return nil, err
	}

	f.mux.Lock()
	f.checksum = sha256.Sum256(b)
	f.data.Reset()

	_, err = f.data.Write(b)
	if err != nil {
		f.mux.Unlock()
		return nil, err
	}
	f.mux.Unlock()

	return b, nil
}

// Checksum prepares a SHA256 checksum of the file contents
func (f *File) Checksum() ([32]byte, error) {
	b, err := f.Read()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(b), nil
}

// Status returns the RPC format for the File
func (f *File) Status() *zync.File {
	return &zync.File{
		Cid:          f.CID.String(),
		AbsolutePath: f.AbsolutePath.String(),
	}
}

func (f *File) attachWatcher(w *Watcher) {
	f.mux.Lock()
	f.Watcher = w
	f.mux.Unlock()
}

func (f *File) detachWatcher() {
	f.mux.Lock()
	f.Watcher = nil
	f.mux.Unlock()
}
