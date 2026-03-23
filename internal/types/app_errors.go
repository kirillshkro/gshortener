package types

import (
	"errors"
)

var (
	ErrEmptyParams = errors.New("empty params")
	ErrFileOpen    = errors.New("file not opened")
	ErrNotFound    = errors.New("key not found")
)

type ErrUnique struct {
	CauseURL RawURL
	ShortURL ShortURL
	Err      error
}

func (e *ErrUnique) Error() string {
	return "Field with key value: " + string(e.CauseURL) + " already exists"
}

func (e *ErrUnique) Unwrap() error {
	return e.Err
}
