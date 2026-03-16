package types

import "errors"

var (
	ErrEmptyParams = errors.New("empty params")
	ErrFileOpen    = errors.New("file not opened")
)

type ErrDuplicateKey struct {
	key string
}

func (e *ErrDuplicateKey) Error() string {
	return "key " + e.key + " already exists"
}
