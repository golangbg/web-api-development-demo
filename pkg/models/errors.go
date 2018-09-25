package models

import "fmt"

// ValidationError is a custom error type for passing validation errors which implements the Error interface
// https://gobyexample.com/errors
type ValidationError struct {
	Field string
	Err   string
}

// Error returns a string which represents the full error messages
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Err)
}
