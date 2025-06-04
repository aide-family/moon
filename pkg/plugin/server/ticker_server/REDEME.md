# 基于Kratos框架的可管理定时任务系统设计与实现

## 背景
在微服务架构中，定时任务管理是一个常见且重要的需求。本文介绍一种基于Kratos框架的可管理定时任务系统设计方案，该方案支持：

- 多任务并行管理
- 动态增删定时任务
- 统一生命周期管理
- 完善的错误处理机制

## 设计目标
1. 模块化：与Kratos框架深度集成，实现transport.Server接口
2. 动态管理：支持运行时动态添加/移除定时任务
3. 资源复用：实现ID回收机制避免内存泄漏
4. 安全可靠：完善的超时控制和错误处理

## 核心结构解析

### 1. 基础定时器 Ticker
```go
type Ticker struct {
    interval time.Duration
    ticker   *time.Ticker
    stop     chan struct{}
    task     *TickTask
    helper   *log.Helper
}

type TickTask struct {
    Fn      func(ctx context.Context, isStop bool) error
    Name    string
    Timeout time.Duration
}
```

特点：

- 内置超时控制（默认10秒）

```go
func (t *Ticker) call(ctx context.Context) {
	timeout := t.task.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := t.task.Fn(ctx, false); err != nil {
		t.helper.Errorf("execute task %s error: %v", t.task.Name, err)
	}
}
```

- 支持优雅停止回调

```go
func (t *Ticker) Start(ctx context.Context) error {
    t.ticker = time.NewTicker(t.interval)
    go func() {
        for {
            select {
            case <-t.ticker.C:
                t.call(ctx)
            case <-t.stop:
                return
            case <-ctx.Done():
                return
            }
        }
    }()
    return nil
}

func (t *Ticker) Stop(ctx context.Context) error {
    close(t.stop)
    t.ticker.Stop()
    if err := t.task.Fn(ctx, true); err != nil {
        t.helper.Errorf("execute task %s error: %v", t.task.Name, err)
    }
    return nil
}
```

- 集成日志系统

```go
func WithTickerLogger(logger log.Logger) TickerOption {
    return func(t *Ticker) {
        t.helper = log.NewHelper(log.With(logger, "module", "server.tick"))
    }
}
```

- 使用场景

```go
// TestNewTicker verifies that TestNewTicker correctly initializes a Ticker with the given interval and task.
func TestNewTicker(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    interval := 1 * time.Second
    start := time.Now()
    task := &server.TickTask{
        Fn: func(ctx context.Context, isStop bool) error {
            if isStop {
                t.Logf("Task stopped")
                return nil
            }
            diff := time.Now().Sub(start)
            diff = diff.Round(time.Second)
            if diff < interval {
                t.Errorf("Expected task to be executed after %v, but it was executed after %v", interval, diff)
                return fmt.Errorf("task executed after %v", diff)
            }
            t.Logf("Task executed after %v", diff)
            return nil
        },
        Name:    "定时器",
        Timeout: 0,
    }

    ticker := server.NewTicker(interval, task)
    err := ticker.Start(ctx)
    if err != nil {
        t.Fatalf("Failed to start timer: %v", err)
    }

    <-ctx.Done()
    ticker.Stop(ctx)
}
```

- 在kratos框架中使用

```go
func NewTicker(bc *conf.Bootstrap, healthService *service.HealthService, logger log.Logger) *server.Ticker {
    serverConfig := bc.GetServer()
    microConfig := bc.GetPalace()
    return server.NewTicker(serverConfig.GetOnlineInterval().AsDuration(), &server.TickTask{
        Name:    "health.Online",
        Timeout: microConfig.GetTimeout().AsDuration(),
        Fn: func(ctx context.Context, isStop bool) error {
            if isStop {
                return healthService.Offline(ctx)
            }
            return healthService.Online(ctx)
        },
    }, server.WithTickerLogger(logger))
}
```

### 2. 定时器管理器 Tickers

```go
type Tickers struct {
    mu      sync.RWMutex
    autoID  uint64
    recycle []uint64
    tickers map[uint64]*Ticker
    logger  log.Logger
}
```

特性：

- 线程安全设计（RWMutex）
- ID自动生成与回收机制
- 支持批量任务初始化
- 统一生命周期管理

**核心方法详解**

- 动态ID管理机制

```go
func (t *Tickers) Add(interval time.Duration, task *TickTask) uint64 {
    t.mu.Lock()
    defer t.mu.Unlock()
    id := t.autoID
    if len(t.recycle) > 0 {
        id = t.recycle[0]
        t.recycle = t.recycle[1:]
    } else {
        t.autoID++
    }
    ticker := NewTicker(interval, task, WithTickerLogger(t.logger))
    defer ticker.Start(context.Background())
    t.tickers[id] = ticker
    return id
}
```

- 动态ID回收机制

```go
func (t *Tickers) Remove(id uint64) {
    t.mu.Lock()
    defer t.mu.Unlock()
    ticker, ok := t.tickers[id]
    if !ok {
        return
    }
    ticker.Stop(context.Background())
    delete(t.tickers, id)
    t.recycle = append(t.recycle, id)
}
```

- 优雅启停

```go
func (t *Tickers) Start(ctx context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, ticker := range t.tickers {
		ticker.Start(ctx)
	}
	return nil
}

func (t *Tickers) Stop(ctx context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, ticker := range t.tickers {
		ticker.Stop(ctx)
	}
	return nil
}
```

- 批量任务管理

```go
func WithTickersTasks(tasks map[time.Duration]*TickTask) TickersOption {
    return func(t *Tickers) {
        for interval, task := range tasks {
            t.Add(interval, task)
        }
    }
}
```

- 使用场景

```go
func TestTestNewTickers(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    list := []time.Duration{
        1 * time.Second,
        2 * time.Second,
        3 * time.Second,
        4 * time.Second,
        5 * time.Second,
    }
    start := time.Now()
    task := make(map[time.Duration]*server.TickTask)
    for _, v := range list {
        task[v] = &server.TickTask{
            Fn: func(ctx context.Context, isStop bool) error {
                if isStop {
                    t.Logf("Task stopped")
                    return nil
                }
                diff := time.Now().Sub(start)
                diff = diff.Round(time.Second)
                if diff < v {
                    t.Errorf("Expected task to be executed after %v, but it was executed after %v", v, diff)
                    return fmt.Errorf("task executed after %v: %v", v, diff)
                }
                t.Logf("Task executed after %v: %v", v, diff)
                return nil
            },
            Name:    fmt.Sprintf("%v", v),
            Timeout: 0,
        }
    }

    tickers := server.NewTickers(server.WithTickersTasks(task))
    err := tickers.Start(ctx)
    if err != nil {
        t.Fatalf("Failed to start timer: %v", err)
    }

    tickers.Add(1*time.Second, &server.TickTask{
        Fn: func(ctx context.Context, isStop bool) error {
            t.Logf("Add 1s Task executed")
            return nil
        },
        Name:    "1s",
        Timeout: 0,
    })

    tickers.Add(2*time.Second, &server.TickTask{
        Fn: func(ctx context.Context, isStop bool) error {
            t.Logf("Add 2s Task executed")
            return nil
        },
        Name:    "2s",
        Timeout: 0,
    })

    <-ctx.Done()
    tickers.Stop(ctx)
}
```

