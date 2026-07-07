// Package types provides custom error types and their methods.
package types

import (
	"errors"
)

var (
	// ErrEmptyParams is returned when the input parameters are empty.
	ErrEmptyParams = errors.New("empty params")
	// ErrFileOpen is returned when a file cannot be opened.
	ErrFileOpen = errors.New("file not opened")
	// ErrNotFound is returned when a key is not found in the data store.
	ErrNotFound = errors.New("key not found")
	// ErrInvalidArgument is returned when an invalid argument is provided.
	ErrInvalidArgument = errors.New("invalid argument")
)

// ErrUnique represents an error where a field with a certain key value already exists.
type ErrUnique struct {
	// CauseURL is the URL that caused the error.
	CauseURL RawURL
	// ShortURL is the shortened URL associated with the cause URL.
	ShortURL ShortURL
	// Err holds the underlying error if any.
	Err error
}

// Error returns a string representation of the error.
func (e *ErrUnique) Error() string {
	return "Field with key value: " + string(e.CauseURL) + " already exists"
}

// Unwrap returns the underlying error if any.
func (e *ErrUnique) Unwrap() error {
	return e.Err
}
