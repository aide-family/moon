package mq

import (
	"strconv"
	"testing"
	"time"
)

func TestNewMockMQ(t *testing.T) {
	mq := NewMockMQ()
	defer mq.Close()

	ch := mq.Receive("test")
	go func() {
		for msg := range ch {
			t.Logf("receive message: %s", msg.Data)
		}
		t.Log("receiver exit")
	}()
	for i := 0; i < 10; i++ {
		mq.Send("test", []byte("hello world "+strconv.Itoa(i)))
		time.Sleep(time.Second)
		if i == 5 {
			mq.RemoveReceiver("test")
		}
	}
}
