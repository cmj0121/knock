package progress

import (
	"fmt"
	"os"
	"strings"
)

type WriterStater interface {
	Write(p []byte) (n int, err error)
	Stat() (os.FileInfo, error)
}

type Progress struct {
	*Spinner
	WriterStater

	width int
}

func New() *Progress {
	return &Progress{
		Spinner:      NewSpinner(),
		WriterStater: os.Stderr,
		width:        48,
	}
}

// update progress
func (p *Progress) AddProgress(msg string, args ...interface{}) {
	switch p.IsTerminal() {
	case true:
		dots := strings.Repeat(".", p.width)
		fmt.Fprintf(p.WriterStater, "\x1b[s\x1b[2K[%v] %v %v\x1b[u", p.Spinner, dots, fmt.Sprintf(msg, args...))
	}
}

// add text result
func (p *Progress) AddText(msg string, args ...interface{}) {
	switch p.IsTerminal() {
	case true:
		fmt.Fprintf(p.WriterStater, "\x1b[2K[+] %v\n", fmt.Sprintf(msg, args...))
	case false:
		fmt.Fprintf(p.WriterStater, "[+] %v\n", fmt.Sprintf(msg, args...))
	}
}

// add an error message
func (p *Progress) AddError(err error) {
	switch p.IsTerminal() {
	case true:
		fmt.Fprintf(p.WriterStater, "\x1b[2K[!] %v\n", err)
	case false:
		fmt.Fprintf(p.WriterStater, "[!] %v\n", err)
	}
}

func (p *Progress) IsTerminal() bool {
	st, _ := p.WriterStater.Stat()
	return (st.Mode() & os.ModeCharDevice) == os.ModeCharDevice
}
