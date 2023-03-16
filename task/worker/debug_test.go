package worker

import (
	"github.com/cmj0121/knock/task/producer"
)

func ExampleDebug() {
	producer, _ := producer.NewCIDRProducer("127.0.0.1/30")

	debug := Debug{}
	debug.Open()        // nolint
	defer debug.Close() // nolint

	debug.Run(producer.Produce(0)) // nolint
	// Output:
	// 127.0.0.0
	// 127.0.0.1
	// 127.0.0.2
	// 127.0.0.3
}
