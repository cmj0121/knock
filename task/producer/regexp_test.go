package producer

import (
	"fmt"
)

func ExampleRegexpProducer() {
	producer, _ := NewRegexpProducer(`\d`)
	for word := range producer.Produce(0) {
		fmt.Println(word)
	}

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
}
