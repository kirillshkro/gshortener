package audit

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

var (
	auditService *FileAuditService
	once         sync.Once
)

type FileAuditService struct {
	file *os.File
}

func newAuditService(filename string) (*FileAuditService, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileAuditService{file: file}, nil
}

func (a *FileAuditService) Notify(e types.Event) {
	if err := json.NewEncoder(a.file).Encode(e); err != nil {
		log.Println("Could't write event to file ", a.file.Name())
	}
}

func GetAuditService(filename string) (*FileAuditService, error) {
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

type NetAuditService struct {
	url    string
	client *http.Client
}

func newNetAuditService(url string) (*NetAuditService, error) {
	return &NetAuditService{url: url,
		client: &http.Client{},
	}, nil
}

func (a *NetAuditService) Notify(e types.Event) {
	bodyReq, _ := json.Marshal(e)
	if _, err := a.client.Post(a.url, "application/json", bytes.NewBuffer(bodyReq)); err != nil {
		log.Println("Error sending audit event:", err)
	}
}
