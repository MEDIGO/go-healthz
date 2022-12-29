package healthz

import "fmt"

// Warning is a type of error that is considered a warning instead of a failure.
// It will not cause the health check to fail, but the warning will appear
// in the JSON.
type Warning struct {
	msg string
}

func (w Warning) Error() string {
	return w.msg
}

// Warn returns a Warning with given message
func Warn(msg string) error {
	return Warning{msg: msg}
}

// Warnf formats a Warning
func Warnf(format string, args ...interface{}) error {
	return Warning{msg: fmt.Sprintf(format, args...)}
}

// IsWarning returns true if the error is a Warning instead of a failure.
func IsWarning(err error) bool {
	_, ok := err.(Warning)
	return ok
}
