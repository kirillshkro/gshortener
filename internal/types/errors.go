package types

import "errors"

var (
	ErrFileNotOpened = errors.New("file not opened")
	ErrEmptyValues   = errors.New("empty values")
)
