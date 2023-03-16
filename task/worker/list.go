package worker

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func init() {
	worker := ListTasks{}
	Register(worker)
}

// the debugger worker and just show the word in STDOUT
type ListTasks struct {
}

// the unique name of worker
func (ListTasks) Name() string {
	return "list"
}

// show the help message
func (ListTasks) Help() string {
	return "list all the tasks"
}

// the dummy open method
func (ListTasks) Open() (err error) {
	log.Debug().Msg("dummy open")
	return
}

// the dummy close method
func (ListTasks) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (ListTasks) Run(producer <-chan string) (err error) {
	worker_lock.Lock()
	defer worker_lock.Unlock()

	for _, worker := range worker_pool {
		// show the worker info
		fmt.Printf("%-16v %v\n", worker.Name(), worker.Help())
	}

	return
}

// copy the current worker settings and generate a new instance
func (ListTasks) Dup() (worker Worker) {
	worker = ListTasks{}
	return
}
