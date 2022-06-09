package internal

import "fmt"

type NarwhalError struct {
	Op  string
	Msg string
}

func (err *NarwhalError) Error() string {
	return fmt.Sprintf("ERROR %s\n%s", err.Op, err.Msg)
}

func NewError(op, msg string) *NarwhalError {
	err := new(NarwhalError)
	err.Op = op
	err.Msg = msg
	return err
}
