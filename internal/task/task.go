package task

import (
	"fmt"
	"sync"
)

var (
	// the internal task lock used when register task into pool
	task_lock = sync.Mutex{}
	// the named task pool
	task_pool = map[string]Task{}
)

// the interface of the task executed by the knock.
type Task interface {
	// the name of the task, should be unique in the task pools
	Name() string

	// run the necessary prepared actions before executed
	Prologue(ctx *Context)

	// run the necessary clean-up actions after task finished
	Epilogue(ctx *Context)

	// execute the task with passed context, return error when fail
	Execute(ctx *Context) error
}

// register task which the name of task should be unique, or raise panic
func RegisterTask(task Task) {
	task_lock.Lock()
	defer task_lock.Unlock()

	name := task.Name()
	if _, ok := task_pool[name]; ok {
		// already register, panic
		panic(fmt.Sprintf("duplicate task name: %v", name))
	}

	task_pool[name] = task
}

// get task by name, may return nil
func GetTask(name string) (task Task, ok bool) {
	task_lock.Lock()
	defer task_lock.Unlock()

	task, ok = task_pool[name]
	return
}
