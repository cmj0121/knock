package producer

import (
	"fmt"
)

func ExampleCIDRProducer() {
	producer, _ := NewCIDRProducer("127.0.0.1/30")
	for word := range producer.Produce(0) {
		fmt.Println(word)
	}

	// Output:
	// 127.0.0.0
	// 127.0.0.1
	// 127.0.0.2
	// 127.0.0.3
}
