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

var ErrUknokwnCompType = errors.New("unknown compression format")

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
		return nil, ErrUknokwnCompType
	}

	return &compReader{
		r:        r,
		compType: compType,
		zr:       zr,
	}, nil
}

func (c compReader) Read(b []byte) (n int, err error) {
	return c.zr.Read(b)
}

func (c *compReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
