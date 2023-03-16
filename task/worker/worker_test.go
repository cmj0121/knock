package worker

import (
	"testing"

	"github.com/cmj0121/knock/task/producer"
)

func TestWorker(t *testing.T) {
	testWorker := func(name string) func(t *testing.T) {
		return func(t *testing.T) {
			producer, _ := producer.NewCIDRProducer("127.0.0.1/30")
			worker, ok := GetWorker(name)

			if !ok {
				// cannot get the worker by name
				t.Fatalf("cannot get worker %v (%v)", name, worker)
			}

			if err := worker.Open(); err != nil {
				// cannot allocate resource
				t.Fatalf("cannot open worker %v: %v", name, err)
			}
			defer func() {
				if err := worker.Close(); err != nil {
					// cannot close allocated resource
					t.Fatalf("cannot close worker %v: %v", name, err)
				}
			}()

			if err := worker.Run(producer.Produce(0)); err != nil {
				// cannot run the worker
				t.Errorf("cannot run worker %v: %v", name, err)
			}
		}
	}

	t.Run("debug", testWorker("debug"))
}
