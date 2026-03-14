package storage

import (
	"context"
	"database/sql"
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
	_ "github.com/lib/pq"
)

type DBStorage struct {
	db *sql.DB
}

var (
	dbinstance *DBStorage
	dbonce     sync.Once
)

func (s *DBStorage) Data(key types.ShortURL) (types.RawURL, error) {
	var originalUrl types.RawURL
	if key != "" {
		if row := s.db.QueryRowContext(context.Background(), "select original_url from urls where short_url = $1", key); row != nil {
			if err := row.Scan(&originalUrl); err != nil {
				return "", err
			}
			return originalUrl, nil
		}
	}
	return "", types.ErrEmptyParams
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
		err error
	)
	dbonce.Do(func() {
		dbinstance, err = newDBStorage(conn)
		if err != nil {
			return
		}
		if _, err = dbinstance.db.ExecContext(context.Background(),
			"create table urls (id serial primary key, short_url text not null, original_url text not null);"); err != nil {
			return
		}
		if _, err = dbinstance.db.ExecContext(context.Background(), "create unique index original_url_idx on urls (original_url);"); err != nil {
			return
		}
		if _, err = dbinstance.db.ExecContext(context.Background(), "create unique index short_url_idx on urls (short_url);"); err != nil {
			return
		}
	})
	return dbinstance, err
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
