package internal

import (
	"fmt"
	"strings"
)

type NarwhalError struct {
	Op  string
	Msg string
}

func (err *NarwhalError) Error() string {
	return fmt.Sprintf("ERROR %s\n%s", err.Op, err.Msg)
}

type TransferConnNotExist struct{}

func (err *TransferConnNotExist) Error() string {
	return "Transfer connection not exist, maybe closed"
}

func IsConnClosed(err error) bool {
	return strings.Contains(err.Error(), "Transfer connection not exist, maybe closed")
}

func IsPortInUsed(err error) bool {
	return strings.Contains(err.Error(), "address already in use")
}

func NewError(op, msg string) *NarwhalError {
	err := new(NarwhalError)
	err.Op = op
	err.Msg = msg
	return err
}
