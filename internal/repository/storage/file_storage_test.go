package storage

import (
	"os"
	"strconv"
	"testing"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/internal/types/model"
	"github.com/stretchr/testify/suite"
)

type storageOriginalURLSuite struct {
	suite.Suite
	fs *FileStorage
}

func (s *storageOriginalURLSuite) SetupSuite() {
	f, err := GetFileStorage("test.json")
	s.Assert().NoError(err)
	s.fs = f
}

func (s *storageOriginalURLSuite) TearDownSuite() {
	s.fs.Close()
	os.Remove("test.json")
}

func (s *storageOriginalURLSuite) Test_StorageOriginalURL() {
	for i := range 2 {
		ss := strconv.Itoa(i)
		_, _, err := s.fs.OriginalURL("testx" + types.ShortURL(ss))
		s.Require().NoError(err)
	}
}

func (s *storageOriginalURLSuite) Test_StorageCreate() {
	err := s.fs.Create(model.URLData{
		ShortURL:    "test0",
		OriginalURL: "testx0",
	})
	s.Require().NoError(err)
	err = s.fs.Create(model.URLData{
		ShortURL:    "test1",
		OriginalURL: "testx1",
	})
	s.Require().NoError(err)
	err = s.fs.Create(model.URLData{
		ShortURL:    "test2",
		OriginalURL: "testx2",
	})
	s.Require().NoError(err)
	err = s.fs.Create(model.URLData{
		ShortURL:    "test3",
		OriginalURL: "testx3",
	})
	s.Require().NoError(err)
}

func (s *storageOriginalURLSuite) Test_StorageGetCounter() {
	counter, err := s.fs.GetCounter()
	s.Require().NoError(err)
	s.Assert().Greater(counter, int64(0))
}

func (s *storageOriginalURLSuite) Test_GetFileStorage() {
	stor, err := GetFileStorage("test.json")
	s.Require().NoError(err)
	s.Assert().NotNil(stor)
	other, err := GetFileStorage("test.json")
	s.Require().NoError(err)
	s.Assert().NotNil(other)
	s.Assert().Equal(stor, other)
}

func Test_FileStorageSuite(t *testing.T) {
	suite.Run(t, new(storageOriginalURLSuite))
}
