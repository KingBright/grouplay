package grouplay

import (
	"fmt"
	"time"
)

type GError struct {
	When time.Time
	What string
}

func NewError(msg string) GError {
	return GError{time.Now(), msg}
}

func (e GError) Error() string {
	return fmt.Sprintf("at %v, %s",
		e.When, e.What)
}
