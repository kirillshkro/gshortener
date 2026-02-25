package shortener

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

func (c *compReader) Close() error {
	return c.zr.Close()
}

func (c compReader) Read(b []byte) (n int, err error) {
	return c.zr.Read(b)
}
