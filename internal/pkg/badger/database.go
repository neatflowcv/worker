package badger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	badgerdb "github.com/dgraph-io/badger/v4"
)

const databaseDirectoryPermission = 0o750

type Database struct {
	db        *badgerdb.DB
	closeOnce sync.Once
	closeErr  error
}

// NewDatabase opens a single Badger handle for a DB directory because the same
// directory must not be opened multiple times. Repositories share this handle,
// and the caller closes it once at the application boundary.
func NewDatabase(dir string) (*Database, error) {
	cleanDir := filepath.Clean(dir)

	err := os.MkdirAll(cleanDir, databaseDirectoryPermission)
	if err != nil {
		return nil, fmt.Errorf("create badger database directory: %w", err)
	}

	options := badgerdb.DefaultOptions(cleanDir)
	options.Logger = nil

	database, err := badgerdb.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open badger database: %w", err)
	}

	return &Database{
		db:        database,
		closeOnce: sync.Once{},
		closeErr:  nil,
	}, nil
}

func (d *Database) Close() error {
	d.closeOnce.Do(func() {
		d.closeErr = d.db.Close()
		if d.closeErr != nil {
			d.closeErr = fmt.Errorf("close badger database: %w", d.closeErr)
		}
	})

	return d.closeErr
}
