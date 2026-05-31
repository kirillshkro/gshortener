package audit

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

var (
	auditService *AuditService
	once         sync.Once
)

type AuditService struct {
	file *os.File
}

func newAuditService(filename string) (*AuditService, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &AuditService{file: file}, nil
}

func (a *AuditService) Notify(e types.Event) error {
	if err := json.NewEncoder(a.file).Encode(e); err != nil {
		return err
	}
	return nil
}

func GetAuditService(filename string) (*AuditService, error) {
	var (
		err error
	)
	once.Do(func() {
		if auditService == nil {
			auditService, err = newAuditService(filename)
		}
	})
	return auditService, err
}
