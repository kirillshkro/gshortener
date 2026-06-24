package middleware

import (
	"compress/gzip"
	"io"
)

type compReader struct {
	r  io.ReadCloser
	zr io.ReadCloser
}

func newCompReader(r io.ReadCloser) (*compReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compReader{
		r:  r,
		zr: zr,
	}, nil
}

// Close closes the underlying gzip reader and the original io.ReadCloser.
// It implements the io.Closer interface.
//
// Returns:
//   - error: An error if closing the underlying readers fails
func (c *compReader) Close() error {
	return c.zr.Close()
}

// Read reads data from the underlying gzip reader into the provided byte slice.
// It implements the io.Reader interface.
//
// Parameters:
//   - b: The byte slice to read data into
//
// Returns:
//   - int: The number of bytes read
//   - error: An error if reading fails
func (c compReader) Read(b []byte) (n int, err error) {
	return c.zr.Read(b)
}
