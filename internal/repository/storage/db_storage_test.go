package storage

import (
	"testing"

	"github.com/kirillshkro/gshortener/internal/mocks"
	"github.com/kirillshkro/gshortener/internal/types"
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

	err := d.mockStorage.SetData(expectedURLData)

	d.Assert().NoError(err)
}

func (d *DBStorageTestSuite) TestDBStorage_SetData_EmptyShortURL() {

	urlData := types.URLData{
		ShortURL:    "",
		OriginalURL: "https://example.com",
	}
	d.mockStorage.EXPECT().SetData(urlData).Return(types.ErrEmptyParams)

	err := d.mockStorage.SetData(urlData)

	d.Assert().Error(err, "ожидалась ошибка при пустом ShortURL")
}

func (d *DBStorageTestSuite) TestDBStorage_SetData_EmptyOriginalURL() {

	urlData := types.URLData{
		ShortURL:    "abc123",
		OriginalURL: "",
	}

	d.mockStorage.EXPECT().SetData(urlData).Return(types.ErrEmptyParams)

	err := d.mockStorage.SetData(urlData)

	d.Assert().Error(err, "ожидалась ошибка при пустом OriginalURL")
}

func (d *DBStorageTestSuite) TestDBStorage_SetData_BothEmpty() {

	urlData := types.URLData{
		ShortURL:    "",
		OriginalURL: "",
	}

	d.mockStorage.EXPECT().SetData(urlData).Return(types.ErrEmptyParams)

	err := d.mockStorage.SetData(urlData)

	d.Assert().Error(err, "ожидалась ошибка при пустых полях")
}

func (d *DBStorageTestSuite) TestDBStorage_Data() {

	shortURL := types.ShortURL("abc123")
	expectedOriginalURL := types.RawURL("https://example.com")
	d.mockStorage.EXPECT().Data(shortURL).Return(expectedOriginalURL, nil)

	originalURL, err := d.mockStorage.Data(shortURL)

	d.Assert().NoError(err)
	d.Assert().Equal(expectedOriginalURL, originalURL)
}

func Test_Main(t *testing.T) {
	suite.Run(t, new(DBStorageTestSuite))
}
