package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kirillshkro/gshortener/internal/mocks"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/internal/types/model"
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

func (d *DBStorageTestSuite) TestDBStorage_Create() {
	// Настройка ожидаемого поведения
	expectedURLOriginalURL := model.URLData{
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
	}
	d.mockStorage.EXPECT().Create(expectedURLOriginalURL).Return(nil)

	err := d.mockStorage.Create(expectedURLOriginalURL)

	d.Assert().NoError(err)
}

func (d *DBStorageTestSuite) TestDBStorage_Create_EmptyShortURL() {

	urlOriginalURL := model.URLData{
		ShortURL:    "",
		OriginalURL: "https://example.com",
	}
	d.mockStorage.EXPECT().Create(urlOriginalURL).Return(types.ErrEmptyParams)

	err := d.mockStorage.Create(urlOriginalURL)

	d.Assert().Error(err, "ожидалась ошибка при пустом ShortURL")
}

func (d *DBStorageTestSuite) TestDBStorage_Create_EmptyOriginalURL() {

	urlOriginalURL := model.URLData{
		ShortURL:    "abc123",
		OriginalURL: "",
	}

	d.mockStorage.EXPECT().Create(urlOriginalURL).Return(types.ErrEmptyParams)

	err := d.mockStorage.Create(urlOriginalURL)

	d.Assert().Error(err, "ожидалась ошибка при пустом OriginalURL")
}

func (d *DBStorageTestSuite) TestDBStorage_Create_BothEmpty() {

	urlOriginalURL := model.URLData{
		ShortURL:    "",
		OriginalURL: "",
	}

	d.mockStorage.EXPECT().Create(urlOriginalURL).Return(types.ErrEmptyParams)

	err := d.mockStorage.Create(urlOriginalURL)

	d.Assert().Error(err, "ожидалась ошибка при пустых полях")
}

func (d *DBStorageTestSuite) TestDBStorage_OriginalURL() {

	shortURL := types.ShortURL("abc123")
	expectedOriginalURL := types.RawURL("https://example.com")
	d.mockStorage.EXPECT().OriginalURL(shortURL).Return(expectedOriginalURL, nil)

	originalURL, err := d.mockStorage.OriginalURL(shortURL)

	d.Assert().NoError(err)
	d.Assert().Equal(expectedOriginalURL, originalURL)
}

func (d *DBStorageTestSuite) TestDBStorage_GetUserURLs() {
	userID := uuid.NewString()
	d.mockStorage.EXPECT().GetUserURLs(userID).Return(
		[]types.UserURL{
			{
				ShortURL:    "http://serv/abc123",
				OriginalURL: "https://example.com/abracadabra",
			},
		}, nil)
	urls, err := d.mockStorage.GetUserURLs(userID)
	d.Assert().NoError(err)
	d.Assert().Positive(len(urls))
}

func (d *DBStorageTestSuite) TestDBStorage_DeleteUserURL() {
	userID := uuid.NewString()
	shortURL := "http://serv/abc123"
	d.mockStorage.EXPECT().DeleteUserURL(userID, types.ShortURL(shortURL)).Return(nil)
	err := d.mockStorage.DeleteUserURL(userID, types.ShortURL(shortURL))
	d.Assert().NoError(err)
}

func Test_Main(t *testing.T) {
	suite.Run(t, new(DBStorageTestSuite))
}
