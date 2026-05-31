package main

import (
	"fmt"
)

// WORK IN PROGRESS / NOT USED FOR NOW

/*
const (
	ErrorGeneric uint8 = iota
	ErrorMediaPipe
	ErrorFFMPEG
	ErrorV4L2
)
*/

type DetailedError struct {
	Where string
	Cause error
}

func (err *DetailedError) Error() string {
	return fmt.Sprintf("[MP +TOAST]: Error %s, %v\n", err.Where, err.Cause)
}

func create_error(where string, err error) error {
	return &DetailedError{
		Where: where,
		Cause: err,
	}
}
