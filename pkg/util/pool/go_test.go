package pool_test

import (
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/util/pool"
)

func TestPool(t *testing.T) {
	poolInstance := pool.NewPool(3, 50)
	poolInstance.Start()

	for i := 0; i < 5; i++ {
		jobID := i
		poolInstance.AddJob(func() {
			log.Debugf("Job %d is running\n", jobID)
			time.Sleep(1 * time.Second)
		})
	}
	go func() {
		time.Sleep(10 * time.Second)
		poolInstance.Stop()
	}()

	poolInstance.Wait()
}

func TestGetPool(t *testing.T) {
	poolInstance := pool.GetWorkerPool()
	poolInstance.Start()

	for i := 0; i < 5; i++ {
		jobID := i
		poolInstance.AddJob(func() {
			log.Debugf("Job %d is running\n", jobID)
			time.Sleep(1 * time.Second)
		})
	}
	go func() {
		time.Sleep(10 * time.Second)
		poolInstance.Stop()
	}()

	poolInstance.Wait()
}

func TestSetPool(t *testing.T) {
	poolInstance := pool.NewPool(30, 50)
	pool.SetWorkerPool(poolInstance)
	poolInstance = pool.GetWorkerPool()
	poolInstance.Start()

	for i := 0; i < 5; i++ {
		jobID := i
		poolInstance.AddJob(func() {
			log.Debugf("Job %d is running\n", jobID)
			time.Sleep(1 * time.Second)
		})
	}
	go func() {
		time.Sleep(10 * time.Second)
		poolInstance.Stop()
	}()

	go func() {
		for i := 0; i < 5; i++ {
			jobID := i
			time.Sleep(1 * time.Second)
			poolInstance.AddJob(func() {
				log.Debugf("waiting Job %d is running\n", jobID)
			})
		}
	}()

	poolInstance.Wait()
}

func TestInit(t *testing.T) {
	pool.SetDefaultJobQueueLen(1)
	pool.SetDefaultWorkerCount(1)
	pool.Init()

	poolInstance := pool.GetWorkerPool()
	poolInstance.Start()
	poolInstance.Start()
	poolInstance.Start()

	for i := 0; i < 5; i++ {
		jobID := i
		poolInstance.AddJob(func() {
			log.Debugf("Job %d is running\n", jobID)
			time.Sleep(1 * time.Second)
		})
	}
	go func() {
		time.Sleep(10 * time.Second)
		poolInstance.Stop()
	}()

	go func() {
		for i := 0; i < 5; i++ {
			jobID := i
			time.Sleep(1 * time.Second)
			poolInstance.AddJob(func() {
				log.Debugf("waiting Job %d is running\n", jobID)
			})
		}
	}()

	poolInstance.Wait()
}
