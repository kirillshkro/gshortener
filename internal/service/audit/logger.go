// Package audit предоставляет реализацию сервисов аудита для логирования событий
// в файловую систему или удаленные HTTP-эндпоинты.
//
// Реализует паттерн Observer для асинхронного уведомления о событиях.
// Поддерживает два типа аудита:
//   - FileAuditService: запись событий в файл
//   - NetAuditService: отправка событий на удаленный сервер
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

// Глобальные переменные для реализации паттерна Singleton.
var (
	auditService    *FileAuditService // Единственный экземпляр файлового аудита
	once            sync.Once         // Синхронизация для однократной инициализации файлового аудита
	netOnce         sync.Once         // Синхронизация для однократной инициализации сетевого аудита
	netAuditService *NetAuditService  // Единственный экземпляр сетевого аудита
)

// FileAuditService представляет сервис аудита, который записывает события в файл.
// Использует JSON-формат для структурированного логирования.
type FileAuditService struct {
	file *os.File // Файловый дескриптор для записи событий
}

// newFileAuditService создает новый экземпляр FileAuditService.
// Если файл существует, открывает его для добавления в конец.
// Если файл не существует, создает новый.
//
// Параметры:
//   - filename: путь к файлу аудита
//
// Возвращает:
//   - *FileAuditService: созданный экземпляр
//   - error: ошибка при открытии или создании файла
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

// Notify записывает событие аудита в файл в формате JSON.
// Является реализацией метода интерфейса наблюдателя.
//
// Параметры:
//   - e: событие для записи
//
// Возвращает:
//   - error: ошибка при сериализации или записи в файл
func (a *FileAuditService) Notify(e *types.Event) error {
	if err := json.NewEncoder(a.file).Encode(e); err != nil {
		log.Println("Could't write event to file ", a.file.Name())
		return err
	}
	return nil
}

// GetAuditService возвращает синглтон-экземпляр FileAuditService.
// Инициализирует сервис при первом вызове.
//
// Параметры:
//   - filename: путь к файлу аудита
//
// Возвращает:
//   - *FileAuditService: экземпляр сервиса
//   - error: ошибка при инициализации
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

// Close закрывает файловый дескриптор аудита.
// Должен быть вызван при завершении работы приложения.
//
// Возвращает:
//   - error: ошибка при закрытии файла
func (a *FileAuditService) Close() error {
	return a.file.Close()
}

// NetAuditService представляет сервис аудита, который отправляет события
// на удаленный HTTP-сервер в формате JSON.
type NetAuditService struct {
	url    string       // URL удаленного сервера аудита
	client *http.Client // HTTP-клиент для отправки запросов
}

// newNetAuditService создает новый экземпляр NetAuditService.
//
// Параметры:
//   - url: адрес удаленного сервера аудита
//
// Возвращает:
//   - *NetAuditService: созданный экземпляр
//   - error: всегда nil в текущей реализации
func newNetAuditService(url string) (*NetAuditService, error) {
	return &NetAuditService{
		url:    url,
		client: &http.Client{},
	}, nil
}

// Notify отправляет событие аудита на удаленный сервер методом POST.
// Сериализует событие в JSON и отправляет с Content-Type: application/json.
//
// Параметры:
//   - e: событие для отправки
//
// Возвращает:
//   - error: ошибка при сериализации или отправке запроса
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

// Close закрывает сетевой аудит-сервис.
// В текущей реализации ничего не делает, так как HTTP-клиент не требует закрытия.
//
// Возвращает:
//   - error: всегда nil
func (a *NetAuditService) Close() error {
	return nil
}

// GetNetAuditService возвращает синглтон-экземпляр NetAuditService.
// Инициализирует сервис при первом вызове.
//
// Параметры:
//   - url: адрес удаленного сервера аудита
//
// Возвращает:
//   - *NetAuditService: экземпляр сервиса
//   - error: ошибка при инициализации
func GetNetAuditService(url string) (*NetAuditService, error) {
	var err error

	netOnce.Do(func() {
		netAuditService, err = newNetAuditService(url)
	})
	return netAuditService, err
}
