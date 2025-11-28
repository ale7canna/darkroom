package rate

import (
	"fmt"
)

type noRateError struct {
	path string
}

func (r noRateError) Error() string {
	return fmt.Sprintf("file %s has no rating", r.path)
}

func noRate(p string) error {
	return noRateError{
		path: p,
	}
}
