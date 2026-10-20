// Package worker 提供轮询 worker 共用的互斥与退避门控。
package worker

import (
	"sync"
	"time"
)

// Gate 保证同一进程内同一时刻只有一个轮询轮次，并在失败后按指数退避。
type Gate struct {
	mu         sync.Mutex
	busy       bool
	notBefore  time.Time
	backoff    time.Duration
	pollEvery  time.Duration
	backoffMax time.Duration
}

// NewGate 创建门控。pollEvery 是正常轮询间隔，backoffMax 是最大退避。
func NewGate(pollEvery, backoffMax time.Duration) *Gate {
	return &Gate{backoff: pollEvery, pollEvery: pollEvery, backoffMax: backoffMax}
}

// Begin 尝试进入一轮。已有轮次或仍在退避期时返回 false。
func (g *Gate) Begin() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.busy || time.Now().Before(g.notBefore) {
		return false
	}
	g.busy = true
	return true
}

// Note 结束一轮。stretch 表示本轮失败，需要延长下次等待时间。
func (g *Gate) Note(stretch bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.busy = false
	wait, next := AdvanceBackoff(g.backoff, g.pollEvery, g.backoffMax, stretch)
	g.backoff = next
	if wait == 0 {
		g.notBefore = time.Time{}
		return
	}
	g.notBefore = time.Now().Add(wait)
}

// AdvanceBackoff 返回本轮应等待的时长和下一轮的退避值。
func AdvanceBackoff(current, pollEvery, backoffMax time.Duration, stretch bool) (time.Duration, time.Duration) {
	if !stretch {
		return 0, pollEvery
	}
	if current < pollEvery {
		current = pollEvery
	}
	next := current * 2
	if next > backoffMax {
		next = backoffMax
	}
	return current, next
}
