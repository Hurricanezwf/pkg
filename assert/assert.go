package assert

import "fmt"

type AssertFunc func() error

func True(condition bool, format string, args ...interface{}) AssertFunc {
	return func() error {
		if !condition {
			return fmt.Errorf(format, args...)
		}
		return nil
	}
}
