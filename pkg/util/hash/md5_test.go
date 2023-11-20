package hash

import (
	"testing"
)

func TestMD5(t *testing.T) {
	t.Log(MD5("hello"), len(MD5("hello")))
	t.Log(MD5("hello1"), len(MD5("1")))
	t.Log(MD5("hello2"), len(MD5("102194912")))
}
