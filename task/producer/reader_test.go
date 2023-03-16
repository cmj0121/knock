package producer

import (
	"fmt"
	"strings"
)

func ExampleReaderProducer() {
	words := "a b c this-is-example test 測試\nテスト\n시험"
	reader := strings.NewReader(words)

	producer := NewReaderProducer(reader)
	for word := range producer.Produce() {
		fmt.Println(word)
	}

	// Output:
	// a
	// b
	// c
	// this-is-example
	// test
	// 測試
	// テスト
	// 시험
}
