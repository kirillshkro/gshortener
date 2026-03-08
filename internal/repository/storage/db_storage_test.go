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
}

func (d *DBTestsSuite) SetupSuite() {
	d.contr = gomock.NewController(d.T())
	d.fakeStorage = mocks.NewMockIStorage(d.contr)
}

func (d *DBTestsSuite) TearDownSuite() {
	d.contr.Finish()
}

func (d *DBTestsSuite) Test_SetData() {
	d.fakeStorage.EXPECT().SetData(gomock.Any()).Return(nil).Times(1)
	data := types.URLData{
		ShortURL:    "test0",
		OriginalURL: "testx0",
	}
	err := d.fakeStorage.SetData(data)
	d.Assert().NoError(err)
}

func Test_DBStorageSuite(t *testing.T) {
	suite.Run(t, new(DBTestsSuite))
}
