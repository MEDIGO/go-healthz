package healthz

import (
	"fmt"
	"sort"
	"strings"
)

// ScopedMultiError contains multiple errors keyed by a unique name
type ScopedMultiError map[string]error

func (e ScopedMultiError) Error() string {
	var sb strings.Builder
	sb.WriteString("multiple errors:")
	var keys []string
	for key := range e {
		keys = append(keys, key)
	}
	sort.Strings(keys) // for deterministic output order
	for _, key := range keys {
		_, _ = fmt.Fprintf(&sb, "\n%s: %v", key, e[key])
	}
	return sb.String()
}

// IsScopedMultiError returns true if the error is a ScopedMultiError.
// It will NOT attempt to unwrap the error.
func IsScopedMultiError(err error) bool {
	_, ok := err.(ScopedMultiError)
	return ok
}
