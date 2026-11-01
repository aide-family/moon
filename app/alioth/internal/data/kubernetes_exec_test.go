package data

import (
	"context"
	"errors"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestPodExecSizeQueue(t *testing.T) {
	queue := newPodExecSizeQueue(80, 24)
	initial := queue.Next()
	if initial == nil || initial.Width != 80 || initial.Height != 24 {
		t.Fatalf("initial size = %+v", initial)
	}
	// 连续缩放只保留最后一帧，避免把中间大小也发给远端。
	queue.Push(100, 30)
	queue.Push(120, 40)
	latest := queue.Next()
	if latest == nil || latest.Width != 120 || latest.Height != 40 {
		t.Fatalf("latest size = %+v", latest)
	}
	queue.Close()
	if size := queue.Next(); size != nil {
		t.Fatalf("closed queue Next = %+v, want nil", size)
	}
}

func TestPodExecOutputDropsWhenFull(t *testing.T) {
	output := newPodExecOutput(1)
	if _, err := output.Write([]byte("a")); err != nil {
		t.Fatalf("write a: %v", err)
	}
	// 没被消费时继续写入只丢新帧，不阻塞调用方。
	if _, err := output.Write([]byte("b")); err != nil {
		t.Fatalf("write b: %v", err)
	}
	if _, err := output.Write([]byte("c")); err != nil {
		t.Fatalf("write c: %v", err)
	}
	first := <-output.data
	if string(first) != "a" {
		t.Fatalf("first chunk = %q, want a", first)
	}
	select {
	case next := <-output.data:
		t.Fatalf("unexpected chunk %q, want dropped", next)
	default:
	}
	output.Close()
	if _, ok := <-output.data; ok {
		t.Fatal("output channel should be closed")
	}
	// 关闭后的写入不算错误，只是丢弃。
	if _, err := output.Write([]byte("d")); err != nil {
		t.Fatalf("write after close: %v", err)
	}
}

func TestPodExecCloseReason(t *testing.T) {
	podResource := schema.GroupResource{Resource: "pods"}
	cases := []struct {
		name     string
		err      error
		canceled error
		want     string
	}{
		{name: "normal exit", want: "exited"},
		{name: "closed", err: errors.New("stream closed"), canceled: context.Canceled, want: "closed"},
		{name: "not found", err: apierrors.NewNotFound(podResource, "web-0"), want: "pod or container not found"},
		{
			name: "forbidden",
			err:  apierrors.NewForbidden(podResource, "web-0", errors.New("no exec")),
			want: "cluster credential is not allowed to exec into this pod",
		},
		{name: "other", err: errors.New("broken pipe"), want: "pod exec session ended"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 日志传 nil，未知错误只按固定文案返回。
			if reason := podExecCloseReason(nil, tc.err, tc.canceled); reason != tc.want {
				t.Fatalf("reason = %q, want %q", reason, tc.want)
			}
		})
	}
}

func TestNormalizePodExecSize(t *testing.T) {
	cases := []struct {
		name               string
		cols, rows         int32
		wantCols, wantRows int32
	}{
		{name: "defaults", wantCols: 80, wantRows: 24},
		{name: "negative defaults", cols: -1, rows: -1, wantCols: 80, wantRows: 24},
		{name: "keeps valid", cols: 120, rows: 40, wantCols: 120, wantRows: 40},
		{name: "clamps", cols: 5000, rows: 5000, wantCols: 1000, wantRows: 1000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cols, rows := normalizePodExecSize(tc.cols, tc.rows)
			if cols != tc.wantCols || rows != tc.wantRows {
				t.Fatalf("size = %dx%d, want %dx%d", cols, rows, tc.wantCols, tc.wantRows)
			}
		})
	}
}
