package storage

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/internal/types/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type URLRepository struct {
	db *gorm.DB

	userRepository UserRepository
}

func (s *URLRepository) OriginalURL(shortURL types.ShortURL) (types.RawURL, error) {
	if shortURL == "" {
		return "", types.ErrEmptyParams
	}
	data, err := gorm.G[model.DataURL](s.db).Select("original_url").Where("short_url = ?", shortURL).First(context.Background())
	if err != nil {
		return "", err
	}
	return data.OriginalURL, nil
}

func (s *URLRepository) Create(reqData model.DataURL) error {
	tx := s.onConflict()
	if err := gorm.G[model.DataURL](tx).Create(context.Background(), &reqData); err != nil {
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

func NewURLRepository(db *gorm.DB) *URLRepository {
	return &URLRepository{
		db:             db,
		userRepository: NewUserRepository(db),
	}
}

func (s *URLRepository) Close() error {
	return nil
}

func (s *URLRepository) shortURL(originalURL types.RawURL) (types.ShortURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	urlOriginalURL, err := gorm.G[model.DataURL](s.db).Where("original_url = ?", originalURL).First(ctx)
	if err != nil {
		return "", err
	}
	return urlOriginalURL.ShortURL, nil
}

func (s *URLRepository) onConflict() *gorm.DB {
	return s.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "original_url"}},
			DoNothing: true,
		},
		clause.Returning{Columns: []clause.Column{{Name: "short_url"}}},
	)
}
