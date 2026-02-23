package shortener

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
)

type compReader struct {
	r        io.ReadCloser
	compType string
	zr       io.ReadCloser
}

func newCompReader(r io.ReadCloser, compType string) (*compReader, error) {
	var (
		zr  io.ReadCloser
		err error
	)
	switch compType {
	case "gzip":
		zr, err = gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
	case "deflate":
		zr, err = zlib.NewReader(r)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Unknown compression format")
	}

	return &compReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compReader) Read(b []byte) (n int, err error) {
	return c.zr.Read(b)
}
