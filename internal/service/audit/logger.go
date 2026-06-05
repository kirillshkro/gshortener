package audit

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

var (
	auditService    *FileAuditService
	once            sync.Once
	netOnce         sync.Once
	netAuditService *NetAuditService
)

type FileAuditService struct {
	file *os.File
}

func newFileAuditService(filename string) (*FileAuditService, error) {
	var file *os.File
	var err error

	// Проверяем, существует ли файл
	if _, err = os.Stat(filename); err == nil {
		// Файл существует, открываем его для записи в конец
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// Файл не существует, создаем новый файл
		file, err = os.Create(filename)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return &FileAuditService{file: file}, nil
}

func (a *FileAuditService) Notify(e *types.Event) error {
	if err := json.NewEncoder(a.file).Encode(e); err != nil {
		log.Println("Could't write event to file ", a.file.Name())
		return err
	}
	return nil
}

func GetAuditService(filename string) (*FileAuditService, error) {
	var (
		err error
	)
	once.Do(func() {
		if auditService == nil {
			auditService, err = newFileAuditService(filename)
		}
	})
	return auditService, err
}

func (a *FileAuditService) Close() error {
	return a.file.Close()
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

func (a *NetAuditService) Notify(e *types.Event) error {
	bodyReq, err := json.Marshal(e)
	if err != nil {
		log.Println("Error marshalling audit event:", err)
		return err
	}
	if _, err := a.client.Post(a.url, "application/json", bytes.NewBuffer(bodyReq)); err != nil {
		log.Println("Error sending audit event:", err)
		return err
	}
	return nil
}

func (a *NetAuditService) Close() error {
	return nil
}

func GetNetAuditService(url string) (*NetAuditService, error) {
	var err error

	netOnce.Do(func() {
		netAuditService, err = newNetAuditService(url)
	})
	return netAuditService, err
}
