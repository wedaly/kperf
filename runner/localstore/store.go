package localstore

import (
	"fmt"
	"io"
)

// Store is a filesystem-like key/value storage.
//
// Each key/value has committed and ingesting status. When OpenWriter returns
// ingestion transcation, the Store opens rootDir/ingesting/$random file to
// receive value data. Once all the data is written, the Commit(ref) moves the
// file into rootDir/committed/ref.
type Store struct {
	rootDir string
}

// NewStore returns new instance of Store.
func NewStore(_rootDir string) *Store {
	return &Store{}
}

// OpenWriter is to initiate a writing operation, ingestion transcation. A
// single ingestion transcation is to open temporary file and allow caller to
// write data into the temporary file. Once all the data is written, the caller
// should call Commit to complete ingestion transcation.
func (s *Store) OpenWriter() (Writer, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// OpenReader is to open committed content named by ref.
func (s *Store) OpenReader(_ref string) (Reader, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// Delete is to delete committed content named by ref.
func (s *Store) Delete(_ref string) error {
	return fmt.Errorf("not implemented yet")
}

// Writer handles writing of content into local store
type Writer interface {
	// Close closes the writer.
	//
	// If the writer has not been committed, this allows aborting.
	// Calling Close on a closed writer will not error.
	io.WriteCloser

	// Commit commits data as file named by ref.
	//
	// Commit always close Writer. If ref already exists, it will return
	// error.
	Commit(ref string) error
}

type Reader interface {
	io.ReaderAt
	io.Closer
	Size() int64
}
