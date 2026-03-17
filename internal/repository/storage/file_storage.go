package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/kirillshkro/gshortener/internal/types"
)

type FileStorage struct {
	file  *os.File
	mu    sync.Mutex
	index map[types.RawURL]bool
	stor  map[types.RawURL]types.ShortURL
}

var (
	instance *FileStorage
	once     sync.Once
)

func GetFileStorage(fPath string) (*FileStorage, error) {
	var err error
	once.Do(func() {
		instance, err = newFileStorage(fPath)
		err = instance.load()
	})
	return instance, err
}

func newFileStorage(fPath string) (*FileStorage, error) {
	file, err := os.OpenFile(fPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file:  file,
		index: make(map[types.RawURL]bool),
		stor:  make(map[types.RawURL]types.ShortURL),
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
func (f *FileStorage) SetData(reqData types.URLData) (err error) {
	var (
		buf []byte
	)
	f.mu.Lock()
	defer f.mu.Unlock()

	key := reqData.ShortURL
	val := reqData.OriginalURL

	if key == "" || val == "" {
		return types.ErrEmptyParams
	}

	if f.keyExist(val) {
		return nil
	}
	item := types.FileData{
		UUID: uuid.New().String(),
		URLData: types.URLData{
			ShortURL:    key,
			OriginalURL: val,
		},
	}

	if buf, err = json.Marshal(item); err != nil {
		return err
	}

	if err := f.appendItem(buf); err != nil {
		return err
	}

	f.index[val] = true
	err = f.file.Sync()
	return err
}

func (f *FileStorage) GetCounter() (counter int64, err error) {
	var info os.FileInfo
	info, err = f.file.Stat()
	if err != nil {
		return 0, err
	}
	counter = 1
	if info.Size() == 0 {
		return counter, nil
	}

	//Перейти ко второй строке файла
	//Первой с записью лога
	if _, err = f.file.Seek(counter+1, io.SeekStart); err != nil {
		return 0, err
	}
	reader := bufio.NewScanner(f.file)
	for reader.Scan() {
		counter++

		if err = reader.Err(); err != nil {
			return 0, err
		}
	}
	//вычитаем последнюю строку файла из счетчика строк
	counter -= 1
	return
}

func (f *FileStorage) AddRecord(r io.Reader) (err error) {
	item := types.FileData{}
	if err = json.NewDecoder(r).Decode(&item); err != nil {
		return err
	}

	rec := types.URLData{
		ShortURL:    item.ShortURL,
		OriginalURL: item.OriginalURL,
	}
	return f.SetData(rec)
}

func (f *FileStorage) keyExist(key types.RawURL) bool {
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

		// Добавляем первую запись
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

func (f *FileStorage) load() (err error) {
	if f.file == nil {
		return types.ErrFileOpen
	}

	var (
		item    types.FileData
		content []types.FileData
		fInfo   os.FileInfo
	)

	fInfo, err = f.file.Stat()
	if err != nil {
		return
	}
	//Если файл только что создан и загружать еще нечего,
	//то ошибки нет
	if fInfo.Size() == 0 {
		return nil
	}

	if err = json.NewDecoder(f.file).Decode(&content); err != nil {
		return
	}

	for _, item = range content {
		f.index[item.OriginalURL] = true
		f.stor[item.OriginalURL] = item.ShortURL
	}
	return
}

func (f *FileStorage) GetShortURL(key types.RawURL) (types.ShortURL, error) {
	if val, ok := f.stor[key]; ok {
		return val, nil
	}
	return "", types.ErrNotFound
}
