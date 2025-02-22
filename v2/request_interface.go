// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Request interface of Command processing golang package.

package command

import "time"

// RequestInterface is commont type of Requesr interface.
type RequestInterface interface {
	// GetVars returns map of request variables.
	GetVars() map[string]string

	// GetData returns request data.
	GetData() []byte

	// SetDate sets date to responce. Used in HTTP request and set custom date
	// to HTTP writer.
	SetDate(date time.Time)
}

// ParseParams parses the input data command parameters.
//
// It attempts to assert the input data to the specified type T.
// If the assertion is successful, it returns the parsed request and a nil error.
// If the assertion fails, it returns the zero value of T and an error indicating
// that the input data was incorrect.
//
// Parameters:
//   - command: a pointer to a CommandData struct representing the command being
//     executed
//   - indata: the input data to be parsed
//
// Returns:
//   - T: the parsed request of type T
//   - error: an error indicating if the input data was incorrect
func ParseParams[T any](indata any) (request T, err error) {

	// Attempt to assert the input data to the specified type T
	// If the assertion fails, set the error variable is set
	request, ok := indata.(T)
	if !ok {
		err = ErrIncorrectInputData
	}

	return
}

// Vars returns map of request variables from input data.
func (c *Commands) Vars(indata any) (map[string]string, error) {
	req, err := ParseParams[RequestInterface](indata)
	if err != nil {
		return nil, err
	}
	return req.GetVars(), nil
}

// Data returns data parameter from input data.
func (c *Commands) Data(indata any) ([]byte, error) {
	req, err := ParseParams[RequestInterface](indata)
	if err != nil {
		return nil, err
	}
	return req.GetData(), nil
}

// SetDate sets date in responce. Used in HTTP request and set custom date to
// HTTP writer.
func (c *Commands) SetDate(indata any, date time.Time) {
	req, err := ParseParams[RequestInterface](indata)
	if err != nil {
		return
	}
	req.SetDate(date)
}
