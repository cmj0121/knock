package task

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cmj0121/knock/task/producer"
	"github.com/cmj0121/knock/task/worker"
	"github.com/rs/zerolog/log"
)

// create a new task manager
func New(name string) (m *TaskManager, err error) {
	w, ok := worker.GetWorker(name)
	if !ok {
		err = fmt.Errorf("cannot get the worker: %v", name)
		return
	}

	m = &TaskManager{
		w: w,
		c: 1,
	}

	return
}

// the task manager which hold the worker and execute by
// number of worker with single producer
type TaskManager struct {
	// the worker instance
	w worker.Worker

	// the number of worker
	c int

	// the durations for producer to generate words
	t time.Duration
}

// change the number of workers
func (m *TaskManager) NumWorkers(c int) (err error) {
	if c <= 0 {
		err = fmt.Errorf("cannot set worker zero or negative: %v", c)
		return
	}

	m.c = c
	return
}

// change the wait durations for producer
func (m *TaskManager) Wait(t time.Duration) (err error) {
	m.t = t
	return
}

// execute the task by the passed producer
func (m *TaskManager) Run(p producer.Producer) (err error) {
	producer := p.Produce(m.t)
	defer p.Close()

	var wg sync.WaitGroup
	// create worker and run via goroutine
	for i := 0; i < m.c; i++ {
		// create the new worker instance, and run with producer
		worker := m.w.Dup()

		if err = worker.Open(); err != nil {
			// cannot allocated worker resource
			return
		}
		defer worker.Close()

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := worker.Run(producer); err != nil {
				log.Error().Err(err).Msg("run worker fail")
			}
		}()

		log.Info().Int("index", i).Msg("create worker")
	}

	done := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		close(done)
	}()

	m.gracefulShutdown(done)
	return
}

func (m *TaskManager) gracefulShutdown(done <-chan struct{}) {
	sigint := make(chan os.Signal, 1)
	defer close(sigint)

	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigint:
		log.Info().Msg("detect SIGINT and start graceful shutdown")
	case <-done:
	}
}
