package types

import (
	"fmt"
	"testing"
	"time"
)

func TestTextJoin(t *testing.T) {
	cnt := 100000
	s := make([]string, 0, cnt)

	for i := 0; i < cnt; i++ {
		s = append(s, "hello")
	}

	ts := time.Now()
	defer t.Log(time.Since(ts))

	TextJoin(s...)
}

type MyStr struct {
	A string `json:"a"`
}

func (m *MyStr) String() string {
	return m.A
}

func TestTextJoinByStringer(t *testing.T) {
	cnt := 100000
	s := make([]fmt.Stringer, 0, cnt)

	for i := 0; i < cnt; i++ {
		s = append(s, &MyStr{A: "hello"})
	}

	ts := time.Now()
	defer t.Log(time.Since(ts))

	TextJoinByStringer(s...)
}

func TestTextJoinByBytes(t *testing.T) {
	cnt := 100000
	s := make([][]byte, 0, cnt)
	for i := 0; i < cnt; i++ {
		s = append(s, []byte("hello"))
	}
	ts := time.Now()
	defer t.Log(time.Since(ts))
	TextJoinByBytes(s...)
}
