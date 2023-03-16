package knock

// the knock instance to run the brute-force task
type Knock struct {
}

// create the Knock instance with the default settings.
func New() (knock *Knock) {
	knock = &Knock{}
	return
}

// run the knock, parse by the passed arguments from CLI and return
// the result.
func (knock *Knock) Run() (exitcode int) {
	exitcode = 0
	return
}
