package storage

import (
	"context"
	"time"

	"github.com/kirillshkro/gshortener/internal/types/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Creator
	Deleter
}

type Creator interface {
	Create(profile model.UserProfile) (int, error)
}

type Deleter interface {
	Delete(id int) error
}

type Reader interface {
	ReadAll(userID int) ([]model.DataURL, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(profile model.UserProfile) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := gorm.G[model.UserProfile](r.db).Create(ctx, &profile); err != nil {
		return 0, err
	}
	return profile.Id, nil
}

func (r *userRepository) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := gorm.G[model.UserProfile](r.db).Where("id = ?", id).Delete(ctx); err != nil {
		return err
	}
	return nil
}
