// Package types provides the set of error types for managing URLs.
package types

// ErrURLDeleted is an error type that indicates a URL has been deleted from the system.
type ErrURLDeleted struct {
	CauseURL RawURL   // The original URL that was deleted.
	ShortURL ShortURL // The shortened URL that corresponds to the deleted original URL.
	Err      error    // The underlying error, if any.
}

// Error returns a string representation of the ErrURLDeleted error.
func (e *ErrURLDeleted) Error() string {
	return "Field with key value: " + string(e.ShortURL) + " cooresponds to " + string(e.CauseURL) + " already deleted"
}

// Unwrap returns the underlying error, if any.
func (e *ErrURLDeleted) Unwrap() error {
	return e.Err
}

type UserIDKey string // A type for representing a user ID as a key in some context.
