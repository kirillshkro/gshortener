package storage

import (
	"database/sql"

	"github.com/kirillshkro/gshortener/internal/types"
)

type DBStorage struct {
	db *sql.DB
}

func (s *DBStorage) Data(key types.ShortURL) (types.RawURL, error) {
	return "", nil
}

func (s *DBStorage) SetData(urlData types.URLData) error {
	return nil
}

func NewDBStorage(conn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DBStorage{
		db: db,
	}, nil
}
