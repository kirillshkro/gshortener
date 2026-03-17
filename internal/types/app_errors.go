package types

import "errors"

var (
	ErrEmptyParams = errors.New("empty params")
	ErrFileOpen    = errors.New("file not opened")
)

type ErrDuplicateKey struct {
	Field string
	Key   string
}

func (e ErrDuplicateKey) Error() string {
	return "Field: " + e.Field + " with key value" + e.Key + " already exists"
}

func NewErrDuplicateKey(field, key string) *ErrDuplicateKey {
	return &ErrDuplicateKey{
		Field: field,
		Key:   key,
	}
}
