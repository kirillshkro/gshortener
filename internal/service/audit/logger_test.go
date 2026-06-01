package audit

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuditTestSuite struct {
	suite.Suite
	auditService *FileAuditService
	netInstance  *NetAuditService
}

func (s *AuditTestSuite) SetupSuite() {
	var (
		err error
	)
	s.auditService, err = GetAuditService("/tmp/audit.log")
	s.Require().NoError(err)
	s.netInstance, err = GetNetAuditService("http://localhost:8080")
	s.Require().NoError(err)
}

func (s *AuditTestSuite) Test_GetFileAuditInstance() {
	var (
		err error
	)
	instance, err := GetAuditService("/tmp/audit.log")
	if s.Assert().NoError(err) {
		s.Assert().Equal(s.auditService, instance)
	}
}

func (s *AuditTestSuite) Test_GetNetAuditInstance() {
	var (
		err error
	)
	instance, err := GetNetAuditService("http://localhost:8080")
	if s.Assert().NoError(err) {
		s.Assert().Equal(s.netInstance, instance)
	}
}

func (s *AuditTestSuite) TearDownSuite() {
	if err := s.auditService.Close(); err != nil {
		s.T().Error(err)
	}
}

func Test_Audit(t *testing.T) {
	suite.Run(t, new(AuditTestSuite))
}
