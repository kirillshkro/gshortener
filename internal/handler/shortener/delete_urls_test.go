package shortener

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/pkg/urlgen"
)

func (s *ServiceTestsSuite) Test_Delete_User_URL_Successfully() {
	//создать несколько сокращенный URL
	url := urlgen.GenerateURL("https://neyandex.me")
	shortURL := Hashing([]byte(url))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(url)))
	// Создаем ссылку для создания сокращенной URL
	recorder := httptest.NewRecorder()
	s.service.URLEncode(recorder, req)
	resp := recorder.Result()
	s.Assert().Equal(http.StatusCreated, resp.StatusCode)
	uCookie := resp.Cookies()[0]
	uID, err := claims.GetUserID(uCookie.Value)
	s.Assert().NoError(err)
	ctx := context.WithValue(req.Context(), types.UserID, uID)
	delURLs := make([]types.ShortURL, 1)
	delURLs = append(delURLs, shortURL)
	rBody, err := json.Marshal(delURLs)
	s.Assert().NoError(err)
	// Отправляем запрос для удаления сокращенной URL
	req = httptest.NewRequestWithContext(ctx, http.MethodDelete, "/api/user/urls", bytes.NewReader(rBody))
	recorder = httptest.NewRecorder()
	s.service.DeleteUserURLs(recorder, req)
	resp = recorder.Result()
	defer resp.Body.Close()
	// Проверка результата
	s.Equal(http.StatusAccepted, resp.StatusCode)
}
