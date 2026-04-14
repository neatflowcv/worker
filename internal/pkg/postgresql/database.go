package postgresql

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const defaultMaxIdleConns = 10
const defaultMaxOpenConns = 20
const defaultConnMaxLifetime = time.Hour

type Database struct {
	db        *gorm.DB
	sqlDB     *sql.DB
	closeOnce sync.Once
	closeErr  error
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := Open(dsn)
	if err != nil {
		return nil, err
	}

	err = ConfigurePool(db)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("extract sql database from gorm database: %w", err)
	}

	return &Database{
		db:        db,
		sqlDB:     sqlDB,
		closeOnce: sync.Once{},
		closeErr:  nil,
	}, nil
}

func Open(dsn string) (*gorm.DB, error) {
	config := new(gorm.Config)
	config.TranslateError = true

	db, err := gorm.Open(
		postgres.Open(dsn),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}

	return db, nil
}

func ConfigurePool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("extract sql database from gorm database: %w", err)
	}

	sqlDB.SetMaxIdleConns(defaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(defaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(defaultConnMaxLifetime)

	return nil
}

func (d *Database) Close() error {
	d.closeOnce.Do(func() {
		d.closeErr = d.sqlDB.Close()
		if d.closeErr != nil {
			d.closeErr = fmt.Errorf("close postgres database: %w", d.closeErr)
		}
	})

	return d.closeErr
}
