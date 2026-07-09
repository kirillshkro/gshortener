// Package storage предоставляет реализации хранилищ для сохранения и получения данных о URL.
// Поддерживаются различные типы хранилищ: база данных (PostgreSQL), файловая система и память.
package storage

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/kirillshkro/gshortener/internal/model"
	"github.com/kirillshkro/gshortener/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// DBStorage представляет реализацию хранилища с использованием базы данных.
// Использует GORM для взаимодействия с PostgreSQL.
type DBStorage struct {
	db *gorm.DB // Экземпляр GORM для работы с БД
}

// Переменные для реализации паттерна Singleton.
var (
	dbinstance *DBStorage // Единственный экземпляр DBStorage
	dbonce     sync.Once  // Синхронизация для однократной инициализации
)

// OriginalURL возвращает оригинальный URL по его короткой версии.
//
// Параметры:
//   - shortURL: короткий идентификатор URL
//
// Возвращает:
//   - types.RawURL: оригинальный URL
//   - error: ошибка, если URL не найден или помечен как удаленный
//
// Ошибки:
//   - types.ErrEmptyParams: если shortURL пустой
//   - types.ErrURLDeleted: если URL был удален пользователем
//   - Другие ошибки GORM при проблемах с БД
func (s *DBStorage) OriginalURL(shortURL types.ShortURL) (types.RawURL, error) {
	if shortURL == "" {
		return "", types.ErrEmptyParams
	}

	data, err := gorm.G[model.URLData](s.db).Select("is_deleted", "original_url").Where("short_url = ?", shortURL).First(context.Background())
	if err != nil {
		return "", err
	}

	// Проверка, не был ли URL удален
	if data.IsDeleted {
		return "", &types.ErrURLDeleted{CauseURL: data.OriginalURL, ShortURL: shortURL, Err: err}
	}

	return data.OriginalURL, nil
}

// Create создает новую запись URL в базе данных.
// При попытке создать дубликат оригинального URL возвращает существующий короткий URL.
//
// Параметры:
//   - reqData: данные для сохранения (оригинальный и короткий URL, UUID пользователя)
//
// Возвращает:
//   - error: ошибка при создании записи или ошибка уникальности
//
// Ошибки:
//   - types.ErrUnique: если URL уже существует, содержит существующий короткий URL
//   - Другие ошибки GORM при проблемах с БД
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

// newDBStorage создает новый экземпляр DBStorage с подключением к базе данных.
// Настраивает логирование GORM с использованием slog.
//
// Параметры:
//   - dsn: строка подключения к PostgreSQL
//
// Возвращает:
//   - *DBStorage: экземпляр хранилища
//   - error: ошибка при подключении к БД
func newDBStorage(dsn string) (*DBStorage, error) {
	// Настройка логгера GORM с использованием slog
	dbLogger := logger.NewSlogLogger(
		slog.New(
			slog.NewJSONHandler(os.Stderr, nil),
		),
		logger.Config{
			LogLevel:             logger.Info,            // Уровень логирования
			SlowThreshold:        500 * time.Millisecond, // Порог медленных запросов
			ParameterizedQueries: true,                   // Параметризованные запросы
			Colorful:             true,                   // Цветной вывод
		},
	)

	conf := &gorm.Config{
		Logger:         dbLogger,
		PrepareStmt:    true, // Подготовка запросов для повышения производительности
		TranslateError: true, // Перевод ошибок БД в понятные типы
	}

	db, err := gorm.Open(postgres.Open(dsn), conf)
	if err != nil {
		return nil, err
	}

	return &DBStorage{
		db: db,
	}, nil
}

// GetDBStorage возвращает синглтон-экземпляр DBStorage.
// Инициализирует подключение к БД при первом вызове и выполняет миграции.
//
// Параметры:
//   - conn: строка подключения к PostgreSQL
//
// Возвращает:
//   - *DBStorage: экземпляр хранилища
//   - error: ошибка при инициализации
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

// populateTables выполняет автоматическую миграцию схемы базы данных.
// Создает таблицы на основе моделей GORM.
//
// Возвращает:
//   - error: ошибка при выполнении миграции
func (s *DBStorage) populateTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.WithContext(ctx).AutoMigrate(&model.URLData{}); err != nil {
		return err
	}
	return nil
}

// Close закрывает соединение с базой данных.
// Для GORM закрытие не требуется, метод реализован для соответствия интерфейсу.
//
// Возвращает:
//   - error: всегда nil
func (s *DBStorage) Close() error {
	return nil
}

// shortURL возвращает короткий URL по оригинальному URL.
// Используется для получения существующего короткого URL при попытке создать дубликат.
//
// Параметры:
//   - originalURL: оригинальный URL
//
// Возвращает:
//   - types.ShortURL: короткий URL
//   - error: ошибка при поиске
func (s *DBStorage) shortURL(originalURL types.RawURL) (types.ShortURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	urlOriginalURL, err := gorm.G[model.URLData](s.db).Where("original_url = ?", originalURL).First(ctx)
	if err != nil {
		return "", err
	}
	return urlOriginalURL.ShortURL, nil
}

// GetUserURLs возвращает все URL, связанные с указанным пользователем.
//
// Параметры:
//   - userUUID: UUID пользователя
//
// Возвращает:
//   - []types.UserURL: список URL пользователя
//   - error: ошибка при получении данных
//
// Ошибки:
//   - types.ErrInvalidArgument: если UUID имеет неверную длину
//   - Другие ошибки GORM при проблемах с БД
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

// DeleteUserURL помечает URL как удаленный для указанного пользователя.
// Использует мягкое удаление (устанавливает флаг is_deleted = true).
//
// Параметры:
//   - ctx: контекст, содержащий user_id для авторизации
//   - shortURL: короткий URL для удаления
//
// Возвращает:
//   - error: ошибка при удалении
//
// Ошибки:
//   - types.ErrInvalidArgument: если userID отсутствует в контексте или имеет неверный формат
//   - Другие ошибки GORM при проблемах с БД
func (s *DBStorage) DeleteUserURL(ctx context.Context, shortURL types.ShortURL) error {
	const (
		uuidLen                   = 36
		userIDKey types.UserIDKey = "user_id"
	)
	var (
		userID string
		ok     bool
	)

	// Извлечение user_id из контекста
	if userID, ok = ctx.Value(userIDKey).(string); !ok {
		return types.ErrInvalidArgument
	}
	if (len(userID) != uuidLen) || len(shortURL) < 3 {
		return types.ErrInvalidArgument
	}

	// Выполнение обновления в транзакции
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[model.URLData](tx).Where("user_uuid = ? AND short_url = ? AND is_deleted = false", userID, shortURL).
			Update(ctx, "is_deleted", true); err != nil {
			return err
		}
		return nil
	})
	return err
}

// onConflict настраивает поведение при конфликте вставки.
// При попытке вставить существующий original_url запрос игнорируется.
// Используется для реализации логики "вставить или ничего не делать".
//
// Возвращает:
//   - *gorm.DB: экземпляр GORM с настроенными Clauses для обработки конфликтов
func (s *DBStorage) onConflict() *gorm.DB {
	return s.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "original_url"}}, // Ключ для проверки конфликта
			DoNothing: true,                                    // Ничего не делать при конфликте
		},
		clause.Returning{Columns: []clause.Column{{Name: "short_url"}}}, // Возвращать короткий URL
	)
}
