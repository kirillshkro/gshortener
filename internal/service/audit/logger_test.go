package audit

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuditTestSuite struct {
	suite.Suite
	auditService *FileAuditService
}

func (s *AuditTestSuite) SetupSuite() {
	var (
		err error
	)
	s.auditService, err = GetAuditService("/tmp/audit.log")
	s.Require().NoError(err)
}

func (s *AuditTestSuite) Test_GetAuditInstance() {
	var (
		err error
	)
	instance, err := GetAuditService("/tmp/audit.log")
	if s.Assert().NoError(err) {
		s.Assert().Equal(s.auditService, instance)
	}
}

func (s *AuditTestSuite) TearDownSuite() {
	if err := s.auditService.file.Close(); err != nil {
		s.T().Error(err)
	}
}

func Test_Audit(t *testing.T) {
	suite.Run(t, new(AuditTestSuite))
}
