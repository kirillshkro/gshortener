package storage

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DBStorage struct {
	db *sqlx.DB
}

func NewDBStorage(connString string) (*DBStorage, error) {
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, err
	}

	return &DBStorage{db: db}, nil
}
