package errors

import "fmt"

type InvalidInputError struct {
	Input any
}

func (err InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input value: %v", err.Input)
}

type InvalidStateError struct {
	Message string
}

func (e InvalidStateError) Error() string {
	return e.Message
}
