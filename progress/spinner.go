package progress

type Spinner struct {
	spinner []string
	index   int
}

func NewSpinner() *Spinner {
	return &Spinner{
		spinner: []string{`-`, `\`, `|`, `/`},
		index:   0,
	}
}

func (s *Spinner) String() (text string) {
	text = s.spinner[s.index]
	s.index = (s.index + 1) % len(s.spinner)
	return
}

func (s *Spinner) Reset() {
	s.index = 0
}
