package storage

import (
	"testing"

	"github.com/kirillshkro/gshortener/internal/mocks"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DBStorageTestSuite struct {
	suite.Suite
	mockStorage *mocks.MockIStorage
	ctrl        *gomock.Controller
}

func (d *DBStorageTestSuite) SetupSuite() {
	d.ctrl = gomock.NewController(d.T())
	d.mockStorage = mocks.NewMockIStorage(d.ctrl)
}

func (d *DBStorageTestSuite) TearDownSuite() {
	d.ctrl.Finish()
}

func (d *DBStorageTestSuite) TestDBStorage_SetData() {
	// Настройка ожидаемого поведения
	expectedURLData := types.URLData{
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
	}
	d.mockStorage.EXPECT().SetData(expectedURLData).Return(nil)

	// Вызов тестируемого метода
	err := d.mockStorage.SetData(expectedURLData)

	// Проверка результата
	d.Assert().NoError(err)
}

func TestDBStorage_SetData_EmptyShortURL(t *testing.T) {
	// Готовим данные
	urlData := types.URLData{
		ShortURL:    "",
		OriginalURL: "https://example.com",
	}

	// Создаем мок
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockIStorage(ctrl)
	mockStorage.EXPECT().SetData(urlData).Return(types.ErrEmptyParams)
	// Вызываем метод
	err := mockStorage.SetData(urlData)

	// Проверяем результат
	assert.Error(t, err, "ожидалась ошибка при пустом ShortURL")
	nilErr := (err == nil)
	assert.False(t, nilErr, "ошибка не должна быть nil")
}

func TestDBStorage_SetData_EmptyOriginalURL(t *testing.T) {
	// Готовим данные
	urlData := types.URLData{
		ShortURL:    "abc123",
		OriginalURL: "",
	}

	// Создаем мок
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockIStorage(ctrl)

	mockStorage.EXPECT().SetData(urlData).Return(types.ErrEmptyParams)

	// Вызываем метод
	err := mockStorage.SetData(urlData)

	// Проверяем результат
	assert.Error(t, err, "ожидалась ошибка при пустом OriginalURL")
	nilErr := (err == nil)
	assert.False(t, nilErr, "ошибка не должна быть nil")
}

func TestDBStorage_SetData_BothEmpty(t *testing.T) {
	// Готовим данные
	urlData := types.URLData{
		ShortURL:    "",
		OriginalURL: "",
	}

	// Создаем мок
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockIStorage(ctrl)

	mockStorage.EXPECT().SetData(urlData).Return(types.ErrEmptyParams)
	// Вызываем метод
	err := mockStorage.SetData(urlData)

	// Проверяем результат
	assert.Error(t, err, "ожидалась ошибка при пустых полях")
	nilErr := (err == nil)
	assert.False(t, nilErr, "ошибка не должна быть nil")
}

func (d *DBStorageTestSuite) TestDBStorage_Data() {
	// Настройка ожидаемого поведения
	shortURL := types.ShortURL("abc123")
	expectedOriginalURL := types.RawURL("https://example.com")
	d.mockStorage.EXPECT().Data(shortURL).Return(expectedOriginalURL, nil)

	// Вызов тестируемого метода
	originalURL, err := d.mockStorage.Data(shortURL)

	// Проверка результата
	d.Assert().NoError(err)
	d.Assert().Equal(expectedOriginalURL, originalURL)
}

func Test_Main(t *testing.T) {
	suite.Run(t, new(DBStorageTestSuite))
}
