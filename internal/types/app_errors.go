package types

import "errors"

var (
	ErrEmptyParams = errors.New("empty params")
	ErrFileOpen    = errors.New("file not opened")
	ErrNotFound    = errors.New("key not found")
)

type ErrUnique struct {
	ShortURL string
	Err      error
}

func (e *ErrUnique) Error() string {
	return "Field with key value" + e.ShortURL + " already exists"
}

func (e *ErrUnique) Unwrap() error {
	return e.Err
}
