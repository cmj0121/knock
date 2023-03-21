package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Progress struct {
	*Spinner
	io.Writer

	width int
}

func New() *Progress {
	return &Progress{
		Spinner: NewSpinner(),
		Writer:  os.Stderr,
		width:   48,
	}
}

// update progress
func (p *Progress) AddProgress(msg string, args ...interface{}) {
	dots := strings.Repeat(".", p.width)
	fmt.Fprintf(p.Writer, "\x1b[s\x1b[2K[%v] %v %v\x1b[u", p.Spinner, dots, fmt.Sprintf(msg, args...))
}

// add text result
func (p *Progress) AddText(msg string, args ...interface{}) {
	fmt.Fprintf(p.Writer, "\x1b[2K[+] %v\n", fmt.Sprintf(msg, args...))
}

// add an error message
func (p *Progress) AddError(err error) {
	fmt.Fprintf(p.Writer, "\x1b[2K[!] %v\n", err)
}
