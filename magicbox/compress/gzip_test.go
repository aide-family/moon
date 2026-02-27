package compress_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/compress"
)

func TestGzip_None(t *testing.T) {
	data := make([]byte, 0)
	reader, err := compress.Gzip(data)
	assert.Nil(t, err)
	assert.NotNil(t, reader)
	uncompressed, err := compress.UnGzip(reader)
	assert.Nil(t, err)
	assert.Equal(t, data, uncompressed)
}

func TestGzip_Some(t *testing.T) {
	data := []byte("hello, world")
	reader, err := compress.Gzip(data)
	assert.Nil(t, err)
	assert.NotNil(t, reader)
	uncompressed, err := compress.UnGzip(reader)
	assert.Nil(t, err)
	assert.Equal(t, data, uncompressed)
}
