package models

import (
	"errors"
	"fmt"
	"strings"
)

type Error struct {
	errors []error
	toErr  error
}

func NewError() *Error {
	return &Error{}
}

func (e *Error) AddError(err error) {
	e.errors = append(e.errors, err)
}

func (e *Error) Errors() []error {
	return e.errors
}

func (e *Error) ToError() error {
	if e.toErr != nil {
		return e.toErr
	}
	if len(e.errors) == 0 {
		return nil
	}
	var sb strings.Builder
	for _, err := range e.errors {
		sb.WriteString(fmt.Sprintf("%v\n", err))
	}
	e.toErr = errors.New(sb.String())
	return e.toErr
}
