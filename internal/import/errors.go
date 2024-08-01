package importer

import "fmt"

type InvalidInputError struct {
	Input any
}

func (err InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input value: %v", err.Input)
}
