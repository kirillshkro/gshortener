package storage

import (
	"testing"

	"github.com/kirillshkro/gshortener/internal/mocks"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DBTestsSuite struct {
	suite.Suite
	fakeStorage *mocks.MockIStorage
	contr       *gomock.Controller
	storage     *Storage
}

func (d *DBTestsSuite) SetupSuite() {
	d.contr = gomock.NewController(d.T())
	d.fakeStorage = mocks.NewMockIStorage(d.contr)
	d.storage = NewStorage()
}

func (d *DBTestsSuite) TearDownSuite() {
	d.contr.Finish()
}

func (d *DBTestsSuite) Test_SetData() {
	data := types.URLData{
		ShortURL:    "test1",
		OriginalURL: "testx1",
	}
	d.fakeStorage.EXPECT().SetData(data).Return(nil).AnyTimes()
	err := d.storage.SetData(data)
	d.Assert().NoError(err)
}

func (d *DBTestsSuite) Test_SetDataWithEmptyShortURL() {
	emptyReq := types.URLData{
		OriginalURL: "",
		ShortURL:    "",
	}
	d.fakeStorage.EXPECT().SetData(emptyReq).Return(types.ErrEmptyValues).Times(1)
	err := d.fakeStorage.SetData(emptyReq)
	d.Assert().Error(err)
}

func (d *DBTestsSuite) Test_GetData() {
	data := types.URLData{
		ShortURL:    "test0",
		OriginalURL: "testx0",
	}
	d.fakeStorage.EXPECT().SetData(data).Return(nil).AnyTimes()
	err := d.fakeStorage.SetData(data)
	d.Assert().NoError(err)
	d.fakeStorage.EXPECT().Data(data.ShortURL).Return(data.OriginalURL, nil).AnyTimes()
	_, err = d.fakeStorage.Data(data.ShortURL)
	d.Assert().NoError(err)
}

func Test_DBStorageSuite(t *testing.T) {
	suite.Run(t, new(DBTestsSuite))
}
