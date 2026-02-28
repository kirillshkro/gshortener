package storage

import (
	"strconv"
	"testing"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StorageData(t *testing.T) {
	fs, err := NewFileStorage("test.json")
	require.NoError(t, err)
	defer fs.Close()

	for i := range 2 {
		ss := strconv.Itoa(i)
		answer, err := fs.Data("testx" + types.ShortURL(ss))
		assert.NotEmpty(t, answer)
		require.NoError(t, err)
	}
}

func Test_STorageSetData(t *testing.T) {
	fs, err := NewFileStorage("test.json")
	require.NoError(t, err)
	defer fs.Close()
	for i := range 2 {
		ss := strconv.Itoa(i)
		err = fs.SetData("testx"+types.ShortURL(ss), "test"+types.RawURL(ss))
		require.NoError(t, err)
	}
}
