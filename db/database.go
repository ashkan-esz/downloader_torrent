package db

import (
	"downloader_torrent/configs"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase() (*Database, error) {
	db, err := gorm.Open(
		postgres.Open(configs.GetConfigs().DbUrl),
		&gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
			TranslateError:         true,
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	return &Database{db: db}, nil
}

func (d *Database) Close() {
	// try not to use it due to gorm connection pooling
	sqlDB, err := d.db.DB()
	if err != nil {
		log.Fatalln(err)
	}
	sqlDB.Close()
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func IsConnectionNotAcceptingError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "57P03"
	}
	return false
}
