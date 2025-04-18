// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command data of Command processing golang package.

package command

import (
	"io"
	"strings"
)

// CommandData represents a command that can be executed by the Commands
// struct. It contains the command name, the input processing types, the
// parameters expected, a description, and a handler function to execute
// the command.
type CommandData struct {
	Cmd       string         // Command name
	ProcessIn ProcessIn      // Input processing types
	Params    string         // Parameters
	Return    string         // Return description
	Descr     string         // Command description
	Request   string         // Request example
	Response  string         // Response example
	Handler   CommandHandler // Command handler
}

// CommandHandler is a function that handles a command.
type CommandHandler func(cmd *CommandData, processIn ProcessIn, data any) (io.Reader, error)

// ParamsSlice returns a slice of parameters from the CommandData struct.
//
// If the CommandData.Params field is empty, it returns nil.
// Otherwise, it splits the CommandData.Params field by "/" and trims
// the resulting parameters of leading and trailing "{}" characters.
// The resulting slice is then returned.
func (c *CommandData) ParamsSlice() []string {
	// If the CommandData.Params field is empty, return nil.
	if c.Params == "" {
		return nil
	}

	// Split and trim parameters in one pass.
	// Count the number of "/" characters in CommandData.Params and
	// add 1 to account for the command name. This is used to initialize
	// the slice with an appropriate capacity.
	params := make([]string, 0, strings.Count(c.Params, "/")+1)
	for _, param := range strings.Split(c.Params, "/") {
		// Trim the parameter of leading and trailing "{}" characters
		// and append to the params slice.
		param = strings.TrimSpace(param)
		param = strings.Trim(param, "{}")
		param = strings.TrimSpace(param)
		params = append(params, param)
	}

	// Return the params slice.
	return params
}
