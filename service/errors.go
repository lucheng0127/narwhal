package service

import "fmt"

// Server Error
type serverError struct {
	msg string
}

type tcpServerError struct {
	msg string
}

func (err *serverError) Error() string {
	return fmt.Sprintf("Narwhal server error %s", err.msg)
}

func (err *tcpServerError) Error() string {
	return fmt.Sprintf("Narwhal TCP server error %s", err.msg)
}

// Client Error
type clientError struct {
	msg string
}

func (err *clientError) Error() string {
	return fmt.Sprintf("Client error %s", err.msg)
}

// Handlers Error
type hRegistryError struct {
	msg string
}

func (err *hRegistryError) Error() string {
	return fmt.Sprintf("Registry client error %s", err.msg)
}

// Socket error
type readError struct {
	msg string
}

func (err *readError) Error() string {
	return fmt.Sprintf("Read data from connection error %s", err.msg)
}

type connNotFound struct {
	msg string
}

func (err *connNotFound) Error() string {
	return fmt.Sprintf("Can't find remote connection for %s", err.msg)
}
