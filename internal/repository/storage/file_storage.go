package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

type FileStorage struct {
	Storage
	file   *os.File
	m      sync.Mutex
	nextID uint
	index  map[types.ShortURL]bool
	stor   types.TStor
}

func NewFileStorage(fPath string) (*FileStorage, error) {
	file, err := os.OpenFile(fPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		Storage: *NewStorage(),
		file:    file,
		index:   make(map[types.ShortURL]bool),
		nextID:  1,
		stor:    make(types.TStor, 0),
	}, nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

/*
Возвращает значение по ключу из файла
*/
func (f *FileStorage) Data(key types.ShortURL) (types.RawURL, error) {
	var (
		fData types.FileData
		err   error
		items []types.FileData
	)

	if info, err := f.file.Stat(); err != nil || info.Size() == 0 {
		return "", err
	}
	if _, err = f.file.Seek(0, 0); err != nil {
		return "", err
	}
	if err = json.NewDecoder(f.file).Decode(&items); err != nil {
		return "", err
	}

	for _, fData = range items {
		if key == fData.ShortURL {
			return fData.OriginalURL, nil
		}
	}
	return "", err
}

/*
Добавляет в файл пару ключ-значение
*/
func (f *FileStorage) SetData(key types.ShortURL, val types.RawURL) error {
	var (
		buf []byte
		err error
	)
	f.m.Lock()
	defer f.m.Unlock()

	if key == "" || val == "" {
		return errors.New("empty params")
	}

	if f.keyExist(key) {
		return errors.New("duplicate key")
	}

	item := types.FileData{
		UUID:        f.nextID,
		ShortURL:    key,
		OriginalURL: val,
	}

	if buf, err = json.Marshal(item); err != nil {
		return err
	}

	if err := f.appendItem(buf); err != nil {
		return err
	}

	f.nextID += 1
	f.index[key] = true
	return nil
}

func (f *FileStorage) keyExist(key types.ShortURL) bool {
	if _, ok := f.index[key]; ok {
		return true
	}
	return false
}

func (f *FileStorage) appendItem(item []byte) error {
	fInfo, err := f.file.Stat()
	if err != nil {
		return err
	}

	if fInfo.Size() == 0 {
		_, err = f.file.WriteString("[\n")
		if err != nil {
			return err
		}

		_, err = f.file.WriteString("  " + string(item) + "\n")
		if err != nil {
			return err
		}
		_, err = f.file.WriteString("]\n")
		return err
	}

	const suffix = "]\n"
	suffixLen := int64(len(suffix))
	if fInfo.Size() < suffixLen {
		return fmt.Errorf("file too small")
	}
	buf := make([]byte, suffixLen)
	_, err = f.file.ReadAt(buf, fInfo.Size()-suffixLen)
	if err != nil {
		return err
	}
	if string(buf) != suffix {
		return fmt.Errorf("file does not end with %q", suffix)
	}

	// Если файл состоит только из "[\n]\n" (пустой массив)
	if fInfo.Size() == 4 { // "[\n]\n" — 4 байта
		// Обрезаем закрывающую скобку
		err = f.file.Truncate(fInfo.Size() - suffixLen)
		if err != nil {
			return err
		}
		// Добавляем первый объект
		_, err = f.file.Seek(0, io.SeekEnd)
		if err != nil {
			return err
		}
		_, err = f.file.WriteString("  " + string(item) + "\n]\n")
		return err
	}

	// Файл содержит как минимум один объект.
	// Находим позицию последнего перевода строки перед суффиксом.
	// Эта позиция находится непосредственно перед суффиксом.
	lastNewlinePos := fInfo.Size() - suffixLen - 1

	// Проверяем, что там действительно перевод строки
	check := make([]byte, 1)
	_, err = f.file.ReadAt(check, lastNewlinePos)
	if err != nil {
		return err
	}
	if check[0] != '\n' {
		return fmt.Errorf("unexpected byte at position %d: expected newline, got %q", lastNewlinePos, check)
	}

	// Обрезаем файл до этой позиции — удаляем последний перевод строки и суффикс
	err = f.file.Truncate(lastNewlinePos)
	if err != nil {
		return err
	}

	// Перемещаем указатель в конец (на позицию после последнего символа предыдущего объекта)
	_, err = f.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// Дописываем запятую, затем новый объект с отступом и закрывающую скобку
	_, err = f.file.WriteString(",\n  " + string(item) + "\n]\n")

	return err
}
