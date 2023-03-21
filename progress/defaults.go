package progress

var (
	progress = New()
)

// update progress
func AddProgress(msg string, args ...interface{}) {
	progress.AddProgress(msg, args...)
}

// add text result
func AddText(msg string, args ...interface{}) {
	progress.AddText(msg, args...)
}

// add an error message
func AddError(err error) {
	progress.AddError(err)
}
