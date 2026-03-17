package storage

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kirillshkro/gshortener/internal/types"
	_ "github.com/lib/pq"
)

type DBStorage struct {
	db *sqlx.DB
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

	if urlData.OriginalURL == "" || urlData.ShortURL == "" {
		return types.ErrEmptyParams
	}

	if _, err := url.Parse(string(urlData.OriginalURL)); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error transaction: %w", err)
	}
	defer tx.Rollback().Error()
	stmt, err := tx.PrepareContext(ctx, "insert into urls (short_url, original_url) values ($1, $2) on conflict (original_url) do nothing")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	result, err := stmt.ExecContext(ctx,
		urlData.ShortURL,
		urlData.OriginalURL)
	if err != nil {
		return fmt.Errorf("error inserting data: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return &types.ErrDuplicateKey{}
	}

	return nil
}

func newDBStorage(conn string) (*DBStorage, error) {
	db, err := sqlx.Open("postgres", conn)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.db.ExecContext(ctx,
		"create table if not exists urls (id serial primary key, short_url text not null, original_url text not null);"); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "create unique index if not exists original_url_idx on urls (original_url);"); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "create unique index if not exists short_url_idx on urls (short_url);"); err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
