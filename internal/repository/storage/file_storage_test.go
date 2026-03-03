package storage

import (
	"strconv"
	"testing"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/suite"
)

type storageDataSuite struct {
	suite.Suite
	fs *FileStorage
}

func (s *storageDataSuite) SetupSuite() {
	f, err := GetFileStorage("test.json")
	s.Assert().NoError(err)
	s.fs = f
}

func (s *storageDataSuite) TearDownSuite() {
	s.fs.Close()
}

func (s *storageDataSuite) Test_StorageData() {
	for i := range 2 {
		ss := strconv.Itoa(i)
		_, err := s.fs.Data("testx" + types.ShortURL(ss))
		s.Require().NoError(err)
	}
}

func (s *storageDataSuite) Test_StorageSetData() {
	for i := range 2 {
		ss := strconv.Itoa(i)
		err := s.fs.SetData("test"+types.RawURL(ss), "testx"+types.ShortURL(ss))
		s.Require().NoError(err)
	}
}

func (s *storageDataSuite) Test_StorageGetCounter() {
	counter, err := s.fs.GetCounter()
	s.Require().NoError(err)
	s.Assert().Greater(counter, int64(0))
}

func (s *storageDataSuite) Test_GetFileStorage() {
	stor, err := GetFileStorage("test.json")
	s.Require().NoError(err)
	s.Assert().NotNil(stor)
	other, err := GetFileStorage("test.json")
	s.Require().NoError(err)
	s.Assert().NotNil(other)
	s.Assert().Equal(stor, other)
}

func Test_FileStorageSuite(t *testing.T) {
	suite.Run(t, new(storageDataSuite))
}
