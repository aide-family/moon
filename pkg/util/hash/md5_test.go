package hash

import (
	"testing"
)

func TestMD5(t *testing.T) {
	t.Log(MD5([]byte("hello")), len(MD5([]byte("hello"))))
	t.Log(MD5([]byte("hello")), len(MD5([]byte("1"))))
	t.Log(MD5([]byte("hello")), len(MD5([]byte("102194912"))))
}
