package storage

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/kirillshkro/gshortener/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type DBStorage struct {
	db *gorm.DB
}

var (
	dbinstance *DBStorage
	dbonce     sync.Once
)

func (s *DBStorage) OriginalURL(shortURL types.ShortURL) (types.RawURL, error) {
	if shortURL == "" {
		return "", types.ErrEmptyParams
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := gorm.G[types.DataURL](s.db).Where("short_url = ?", shortURL).First(ctx)
	if err != nil {
		return "", err
	}
	return data.OriginalURL, nil
}

func (s *DBStorage) Create(reqData types.DataURL) error {
	if err := s.onConflict().Create(&reqData).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			shortURL, err := s.shortURL(reqData.OriginalURL)
			return &types.ErrUnique{
				ShortURL: string(shortURL),
				Err:      err,
			}
		}
		return err
	}
	return nil
}

func newDBStorage(conn string) (*DBStorage, error) {
	dbLogger := logger.NewSlogLogger(
		slog.New(
			slog.NewJSONHandler(os.Stderr, nil),
		),
		logger.Config{
			LogLevel:             logger.Info,
			SlowThreshold:        500 * time.Millisecond,
			ParameterizedQueries: true,
			Colorful:             true,
		},
	)
	conf := &gorm.Config{
		Logger:         dbLogger,
		PrepareStmt:    true,
		TranslateError: true,
	}
	db, err := gorm.Open(postgres.Open(conn), conf)
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
	if err := s.db.WithContext(ctx).AutoMigrate(&types.DataURL{}); err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) Close() error {
	return nil
}

func (s *DBStorage) shortURL(originalURL types.RawURL) (types.ShortURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	urlOriginalURL, err := gorm.G[types.DataURL](s.db).Where("original_url = ?", originalURL).First(ctx)
	if err != nil {
		return "", err
	}
	return urlOriginalURL.ShortURL, nil
}

func (s *DBStorage) onConflict() *gorm.DB {
	return s.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "original_url"}},
			DoNothing: true,
		},
		clause.Returning{Columns: []clause.Column{{Name: "short_url"}}},
	)
}
