package storage

import (
	"github.com/jmoiron/sqlx"
	"github.com/kirillshkro/gshortener/internal/types"
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
	if _, err = db.Exec(`create table if not exists urls (id integer generated always as identity primary key, 
	short_url varchar(255) not null, 
	original_url text not null)`); err != nil {
		return nil, err
	}
	//добавляем индекс для поля short_url
	if _, err = db.Exec(`create unique index if not exists idx_short_url on urls (short_url)`); err != nil {
		return nil, err
	}

	//добавляем ограничение уникальности для поля original_url
	if _, err = db.Exec(`create index if not exists idx_original_url on urls (original_url)`); err != nil {
		return nil, err
	}

	return &DBStorage{db: db}, nil
}

func (s DBStorage) Close() error {
	return s.db.Close()
}

func (s DBStorage) Data(key types.ShortURL) (types.RawURL, error) {
	var result types.RawURL
	if key == "" {
		return "", types.ErrEmptyValues
	}
	if row := s.db.QueryRow(`select original_url from urls where short_url = $1`, key); row != nil {
		if err := row.Scan(&result); err != nil {
			return "", err
		}
	}
	return result, nil
}

func (s DBStorage) SetData(reqData types.URLData) error {
	if reqData.ShortURL == "" || reqData.OriginalURL == "" {
		return types.ErrEmptyValues
	}
	if _, err := s.db.Exec(`insert into urls (short_url, original_url) values ($1, $2)`, reqData.ShortURL, reqData.OriginalURL); err != nil {
		return err
	}
	return nil
}
