package storage

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/internal/types/model"
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
	data, err := gorm.G[model.URLData](s.db).Select("is_deleted", "original_url").Where("short_url = ?", shortURL).First(context.Background())
	if err != nil {
		return "", err
	}

	if data.IsDeleted {
		return "", &types.ErrURLDeleted{CauseURL: data.OriginalURL, ShortURL: shortURL, Err: err}
	}

	return data.OriginalURL, nil
}

func (s *DBStorage) Create(reqData model.URLData) error {
	tx := s.onConflict()
	if err := gorm.G[model.URLData](tx).Create(context.Background(), &reqData); err != nil {
		slog.Error("Current error: " + err.Error())
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			shortURL, err := s.shortURL(reqData.OriginalURL)
			return &types.ErrUnique{
				CauseURL: reqData.OriginalURL,
				ShortURL: shortURL,
				Err:      err,
			}
		}
		return err
	}
	return nil
}

func newDBStorage(dsn string) (*DBStorage, error) {
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
	db, err := gorm.Open(postgres.Open(dsn), conf)
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
	if err := s.db.WithContext(ctx).AutoMigrate(&model.URLData{}); err != nil {
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
	urlOriginalURL, err := gorm.G[model.URLData](s.db).Where("original_url = ?", originalURL).First(ctx)
	if err != nil {
		return "", err
	}
	return urlOriginalURL.ShortURL, nil
}

func (s *DBStorage) GetUserURLs(userUUID string) ([]types.UserURL, error) {
	const uuidLen = 36
	if len(userUUID) != uuidLen {
		return nil, types.ErrInvalidArgument
	}
	urls, err := gorm.G[model.URLData](s.db).Select("short_url", "original_url").Where("user_uuid = ?", userUUID).Find(context.Background())
	if err != nil {
		return nil, err
	}
	var result []types.UserURL
	for _, url := range urls {
		result = append(result, types.UserURL{
			ShortURL:    string(url.ShortURL),
			OriginalURL: string(url.OriginalURL),
		})
	}
	return result, nil
}

func (s *DBStorage) DeleteUserURL(userID string, shortURL types.ShortURL) error {
	const uuidLen = 36
	if (len(userID) != uuidLen) || len(shortURL) < 3 {
		return types.ErrInvalidArgument
	}
	ctx := context.Background()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[model.URLData](tx).Where("user_uuid = ? AND short_url = ? AND is_deleted = false", userID, shortURL).
			Update(ctx, "is_deleted", true); err != nil {
			return err
		}
		return nil
	})
	return err
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
