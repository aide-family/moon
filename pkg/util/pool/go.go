package pool

import (
	"sync"

	"github.com/go-kratos/kratos/v2/log"
)

type Job func()

type Worker struct {
	id       int
	jobQueue chan Job
	quit     chan bool
}

var (
	pool               *Pool
	defaultWorkerCount = 10
	defaultJobQueueLen = 100
	once               sync.Once
)

func init() {
	once.Do(func() {
		if pool == nil {
			pool = NewPool(defaultWorkerCount, defaultJobQueueLen)
		}
	})
}

// GetWorkerPool 获取worker pool
func GetWorkerPool() *Pool {
	return pool
}

// SetWorkerPool 设置worker pool
func SetWorkerPool(p *Pool) {
	pool = p
}

// SetDefaultWorkerCount 设置默认worker数量
func SetDefaultWorkerCount(count int) {
	defaultWorkerCount = count
}

// SetDefaultJobQueueLen 设置默认job queue长度
func SetDefaultJobQueueLen(length int) {
	defaultJobQueueLen = length
}

// Init 初始化协程池
func Init() {
	pool = NewPool(defaultWorkerCount, defaultJobQueueLen)
}

func NewWorker(id int, jobQueue chan Job) *Worker {
	return &Worker{
		id:       id,
		jobQueue: jobQueue,
		quit:     make(chan bool),
	}
}

func (w *Worker) start(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		for {
			select {
			case job := <-w.jobQueue:
				w.callJobFunc(job)
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Worker) callJobFunc(job Job) {
	if job == nil {
		return
	}
	job()
}

func (w *Worker) stop() {
	w.quit <- true
}

type Pool struct {
	jobQueue    chan Job
	workerCount int
	workers     []*Worker
	wg          sync.WaitGroup
	startOnce   sync.Once
}

func NewPool(workerCount, jobQueueLen int) *Pool {
	jobQueue := make(chan Job, jobQueueLen)
	var workers []*Worker
	for i := 0; i < workerCount; i++ {
		worker := NewWorker(i, jobQueue)
		workers = append(workers, worker)
	}
	return &Pool{
		jobQueue:    jobQueue,
		workerCount: workerCount,
		workers:     workers,
	}
}

func (p *Pool) Start() {
	flag := true
	p.startOnce.Do(func() {
		p.wg.Add(p.workerCount)
		for _, worker := range p.workers {
			worker.start(&p.wg)
		}
		log.Info("[Pool] started")
		flag = false
	})
	if flag {
		log.Warn("[Pool] already started")
	}
}

func (p *Pool) AddJob(job Job) {
	p.jobQueue <- job
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Stop() {
	for _, worker := range p.workers {
		worker.stop()
	}
	close(p.jobQueue)
	log.Info("[Pool] stopped")
}
