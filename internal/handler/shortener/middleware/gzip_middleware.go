// Package middleware предоставляет HTTP-промежуточные обработчики (middleware)
// для сжатия данных, логирования, аутентификации и других сквозных задач.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// HandlerWithGzip реализует middleware для сжатия HTTP-трафика.
// Обрабатывает как входящие (распаковка), так и исходящие (упаковка) данные.
//
// Принцип работы:
// 1. Проверяет Content-Encoding запроса и распаковывает тело, если оно сжато
// 2. Проверяет Accept-Encoding клиента для определения поддержки gzip
// 3. Если клиент поддерживает gzip и код ответа < 300, оборачивает ResponseWriter для сжатия
//
// Особенности:
//   - Коды ответа < 300 сжимаются, остальные передаются без сжатия
//   - Автоматически устанавливает Content-Encoding: gzip для сжатых ответов
//   - Поддерживает компрессию как запросов, так и ответов
//
// Параметры:
//   - next: следующий обработчик в цепочке
//
// Возвращает:
//   - http.Handler: обернутый обработчик с поддержкой gzip
func HandlerWithGzip(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Content-Encoding")
		reqGzip := strings.Contains(encoding, "gzip")

		if reqGzip && r.Body != nil {
			gzr, err := newCompReader(r.Body)
			if err != nil {
				http.Error(w, "unknown server error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer gzr.Close()
			r.Body = gzr
		}

		zw := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		respGzip := strings.Contains(acceptEncoding, "gzip")
		if respGzip {
			cw := newCompWriter(w)
			zw = cw
			defer cw.Close()
		}

		next.ServeHTTP(zw, r)
	}
	return http.HandlerFunc(fn)
}

// compWriter реализует http.ResponseWriter с поддержкой сжатия gzip.
// Оборачивает стандартный ResponseWriter и сжимает все записываемые данные.
//
// Важные особенности:
//   - Сжатие применяется только для успешных ответов (код < 300)
//   - Заголовок Content-Encoding устанавливается автоматически
//   - Требуется закрытие для корректной записи сжатых данных
type compWriter struct {
	w  http.ResponseWriter // Оригинальный ResponseWriter
	zw io.WriteCloser      // Gzip-врайтер для сжатия данных
}

// Header возвращает заголовки HTTP-ответа.
// Проксирует вызов к оригинальному ResponseWriter.
//
// Возвращает:
//   - http.Header: заголовки ответа
func (c *compWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает сжатые данные в ответ.
// Данные сжимаются с помощью gzip перед отправкой клиенту.
//
// Параметры:
//   - b: данные для записи
//
// Возвращает:
//   - int: количество записанных байт
//   - error: ошибка при записи
func (c *compWriter) Write(b []byte) (int, error) {
	return c.zw.Write(b)
}

// WriteHeader устанавливает HTTP-статус код и добавляет заголовок Content-Encoding.
// Заголовок Content-Encoding добавляется только для успешных ответов (код < 300).
// Это соответствует лучшим практикам, так как ошибки обычно не сжимаются.
//
// Параметры:
//   - statusCode: HTTP статус код ответа
func (c *compWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip-врайтер и отправляет все буферизированные данные.
// Должна быть вызвана после завершения записи.
//
// Возвращает:
//   - error: ошибка при закрытии gzip-врайтера
func (c *compWriter) Close() error {
	return c.zw.Close()
}

// newCompWriter создает новый обернутый ResponseWriter с поддержкой gzip.
//
// Параметры:
//   - w: оригинальный ResponseWriter
//
// Возвращает:
//   - *compWriter: обернутый ResponseWriter с поддержкой сжатия
func newCompWriter(w http.ResponseWriter) *compWriter {
	return &compWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}
