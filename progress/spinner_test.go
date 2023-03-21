package progress

import (
	"fmt"
)

func ExampleSpinner() {
	spinner := NewSpinner()

	fmt.Println(spinner)
	fmt.Println(spinner)
	fmt.Println(spinner)
	fmt.Println(spinner)
	fmt.Println(spinner)
	// Output:
	// -
	// \
	// |
	// /
	// -
}

func ExampleSpinner_Reset() {
	spinner := NewSpinner()

	fmt.Println(spinner)
	fmt.Println(spinner)
	fmt.Println(spinner)
	spinner.Reset()
	fmt.Println(spinner)
	fmt.Println(spinner)
	fmt.Println(spinner)
	// Output:
	// -
	// \
	// |
	// -
	// \
	// |
}
