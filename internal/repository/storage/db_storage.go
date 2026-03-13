package storage

import (
	"database/sql"
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
	_ "github.com/lib/pq"
	goose "github.com/pressly/goose/v3"
)

type DBStorage struct {
	db *sql.DB
}

var (
	dbinstance *DBStorage
	dbonce     sync.Once
)

func (s *DBStorage) Data(key types.ShortURL) (types.RawURL, error) {
	return "", nil
}

func (s *DBStorage) SetData(urlData types.URLData) error {
	return nil
}

func newDBStorage(conn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return &DBStorage{
		db: db,
	}, nil
}

func GetDBStorage(conn string) (*DBStorage, error) {
	var err error
	dbonce.Do(func() {
		dbinstance, err = newDBStorage(conn)
		err = goose.Up(dbinstance.db, "./migrations")
	})
	return dbinstance, err
}
