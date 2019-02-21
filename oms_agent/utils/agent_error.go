package utils

import "fmt"

type agentError struct {
	arg  int
	prob string
}


func (e *agentError) Error() string {
	return fmt.Sprintf("%d - %s", e.arg, e.prob)
}
