package worker

import (
	"fmt"
	"sync"
)

// the pseudo worker to handle word producer
type Worker interface {
	// the unique name of worker
	Name() string

	// open with necessary resources
	Open() error

	// close the allocated resources
	Close() error

	// execute the worker
	Run(producer <-chan string) error

	// copy the current worker settings and generate a new instance
	Dup() Worker
}

var (
	worker_pool = map[string]Worker{}
	worker_lock = sync.Mutex{}
)

func Register(worker Worker) {
	worker_lock.Lock()
	defer worker_lock.Unlock()

	if _, ok := worker_pool[worker.Name()]; ok {
		err := fmt.Errorf("duplicated worker: %v", worker.Name())
		panic(err)
	}

	worker_pool[worker.Name()] = worker.Dup()
}

func GetWorker(name string) (worker Worker, ok bool) {
	worker_lock.Lock()
	defer worker_lock.Unlock()

	if worker, ok = worker_pool[name]; ok {
		worker = worker.Dup()
	}
	return
}
