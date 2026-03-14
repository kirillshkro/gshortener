package storage

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
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
	var (
		err      error
		provider *goose.Provider
	)
	dbonce.Do(func() {
		var fs fs.FS = os.DirFS("../../../migrations")
		dbinstance, err = newDBStorage(conn)
		if err != nil {
			return
		}
		provider, err = goose.NewProvider(goose.DialectPostgres, dbinstance.db, fs)
		if err != nil {
			return
		}
		defer provider.Close()
		_, err = provider.Up(context.Background())
		if err != nil {
			return
		}
	})
	return dbinstance, err
}
