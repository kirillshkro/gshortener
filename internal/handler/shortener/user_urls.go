// Package shortener предоставляет HTTP-обработчики для работы с сокращением URL.
// Содержит методы для создания, получения и управления короткими ссылками.
package shortener

import (
	"encoding/json"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/types"
)

// Getter определяет интерфейс для получения URL-адресов пользователя.
// Реализует метод GetUserURLs для извлечения всех сокращенных URL,
// связанных с конкретным пользователем.
type Getter interface {
	// GetUserURLs обрабатывает запрос на получение всех URL пользователя.
	// Ожидает, что идентификатор пользователя будет извлечен из контекста запроса.
	//
	// Параметры:
	//   - resp: HTTP-ответ для записи результата
	//   - req: HTTP-запрос с контекстом пользователя
	GetUserURLs(resp http.ResponseWriter, req *http.Request)
}

// GetUserURLs возвращает все сокращенные URL, принадлежащие текущему пользователю.
// Метод извлекает идентификатор пользователя из контекста запроса,
// который должен быть установлен middleware аутентификации.
//
// Процесс обработки:
// 1. Извлечение userID из контекста запроса
// 2. Получение списка URL из хранилища
// 3. Формирование полных коротких URL с добавлением базового адреса
// 4. Возврат результата в формате JSON
//
// Коды ответа:
//   - 200 OK: успешное получение списка URL
//   - 204 No Content: список URL пуст
//   - 401 Unauthorized: ошибка аутентификации или доступа
//
// Параметры:
//   - resp: HTTP-ответ для записи результата
//   - req: HTTP-запрос с контекстом, содержащим идентификатор пользователя
func (s Service) GetUserURLs(resp http.ResponseWriter, req *http.Request) {
	var (
		userID string
		ok     bool
	)

	// Извлечение идентификатора пользователя из контекста запроса
	userID, ok = req.Context().Value(types.UserID).(string)
	if !ok {
		resp.WriteHeader(http.StatusNoContent)
		return
	}

	// Получение списка URL пользователя из хранилища
	urls, err := s.Stor.GetUserURLs(userID)
	if err != nil {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Проверка наличия URL у пользователя
	if len(urls) == 0 {
		resp.WriteHeader(http.StatusNoContent)
		s.createCookie(resp)
		return
	}

	// Формирование полных коротких URL с добавлением базового адреса
	var userURLs []types.UserURL
	for _, url := range urls {
		userURLs = append(userURLs, types.UserURL{
			ShortURL:    string(s.ResultAddr) + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		})
	}

	// Отправка ответа в формате JSON
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(resp).Encode(userURLs); err != nil {
		s.logger.Error("cannot encode response: ", "error: ", err.Error())
		return
	}
}
