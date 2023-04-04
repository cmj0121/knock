package producer

import (
	"fmt"
	"reflect"
	"testing"
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

func TestRegexpProducer(t *testing.T) {
	cases := []struct {
		Pattern string
		Tokens  []string
	}{
		{
			Pattern: `[01]`,
			Tokens:  []string{"0", "1"},
		},
		{
			Pattern: `[0-4]`,
			Tokens:  []string{"0", "1", "2", "3", "4"},
		},
		{
			Pattern: `abc-[0-4]`,
			Tokens:  []string{"abc-0", "abc-1", "abc-2", "abc-3", "abc-4"},
		},
		{
			Pattern: `[ab][0-2]`,
			Tokens:  []string{"a0", "a1", "a2", "b0", "b1", "b2"},
		},
		{
			Pattern: `[ab]{2}`,
			Tokens:  []string{"aa", "ab", "ba", "bb"},
		},
		{
			Pattern: `[ab]{2,3}`,
			Tokens:  []string{"aa", "ab", "ba", "bb", "aaa", "aab", "aba", "abb", "baa", "bab", "bba", "bbb"},
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.Pattern, testRegexpProducer(testcase.Pattern, testcase.Tokens))
	}
}

func testRegexpProducer(pattern string, tokens []string) func(*testing.T) {
	return func(t *testing.T) {
		producer, err := NewRegexpProducer(pattern)
		if err != nil {
			t.Fatalf("cannot generate new regexp producer via %#v: %v", pattern, err)
			return
		}

		ret := []string{}
		for token := range producer.Produce(0) {
			ret = append(ret, token)
		}

		if !reflect.DeepEqual(ret, tokens) {
			t.Errorf("expected %#v got %v: %v", pattern, tokens, ret)
		}
	}
}
