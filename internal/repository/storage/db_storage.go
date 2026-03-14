package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/kirillshkro/gshortener/internal/types"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type DBStorage struct {
	db *sql.DB
}

var (
	dbinstance *DBStorage
	dbonce     sync.Once
)

func (s *DBStorage) Data(key types.ShortURL) (types.RawURL, error) {
	var originalURL types.RawURL
	if key != "" {
		if row := s.db.QueryRowContext(context.Background(), "select original_url from urls where short_url = $1", key); row != nil {
			if err := row.Scan(&originalURL); err != nil {
				return "", err
			}
			return originalURL, nil
		}
	}
	return "", types.ErrEmptyParams
}

func (s *DBStorage) SetData(urlData types.URLData) error {
	if urlData.ShortURL != "" && urlData.OriginalURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := s.db.ExecContext(ctx, "insert into urls (short_url, original_url) values ($1, $2)",
			urlData.ShortURL,
			urlData.OriginalURL); err != nil {
			return err
		}
		return nil
	}
	return types.ErrEmptyParams
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
		err = dbinstance.populateTables()
		if err != nil {
			return
		}
	})
	return dbinstance, err
}

func (s *DBStorage) populateTables() error {
	if err := goose.Up(s.db, "../../../migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
