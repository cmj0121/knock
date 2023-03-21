package progress

import (
	"fmt"
	"os"
)

func ExampleProgress_AddText() {
	progress := New()
	progress.Writer = os.Stdout

	progress.AddText("1")
	progress.AddText("2")
	progress.AddText("3")
	progress.AddText("code %v", 4)
	// Output:
	// [+] 1
	// [+] 2
	// [+] 3
	// [+] code 4
}

func ExampleProgress_AddError() {
	progress := New()
	progress.Writer = os.Stdout

	progress.AddError(fmt.Errorf("1"))
	progress.AddError(fmt.Errorf("2"))
	progress.AddError(fmt.Errorf("3"))
	// Output:
	// [!] 1
	// [!] 2
	// [!] 3
}
