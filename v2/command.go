// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command processing golang package.
package command

import (
	"bytes"
	"fmt"
	"iter"
	"sync"
)

// ErrIncorrectInputData is an error returned when the input data provided is
// incorrect.
var ErrIncorrectInputData = fmt.Errorf("inclorrect input data")
var ErrNotValidChannel = fmt.Errorf("not valid connection channel")

// Commands is a struct that contains a map of command data and a read-write
// mutex for synchronizing access to the map.
type Commands struct {
	m map[string]*CommandData
	*sync.RWMutex
}

// New creates and initializes new Commands object.
func New() *Commands {
	c := &Commands{}
	c.Init()
	return c
}

// Init initialize Commands object.
func (c *Commands) Init() {
	c.m = make(map[string]*CommandData)
	c.RWMutex = new(sync.RWMutex)
}

// Add adds command to commands map.
//
// Parameters:
//   - command: The name of the command.
//   - descr: A short description of the command.
//   - processIn: The type of input processing to use for the command.
//   - params: The parameters expected by the command.
//   - returnDescr: A description of the data returned by the command.
//   - request: The request example.
//   - response: The response example.
//   - handler: The function that handles the command.
//
// Returns:
// - *Commands: The Commands object itself.
func (c *Commands) Add(command, descr string, processIn ProcessIn, params,
	returnDescr, request, response string, handler CommandHandler) *Commands {
	c.Lock()
	defer c.Unlock()

	c.m[command] = &CommandData{
		command, processIn, params, returnDescr, descr, request, response, handler,
	}

	return c
}

// Get returns CommandData from commands map by name.
//
// It returns the CommandData and a boolean exist flag that indicates if the
// command was found.
func (c *Commands) Get(name string) (cmd *CommandData, exists bool) {
	c.RLock()
	defer c.RUnlock()

	cmd, exists = c.m[name]
	return
}

// Del removes command from commands map.
func (c *Commands) Del(name string) {
	c.Lock()
	delete(c.m, name)
	c.Unlock()
}

// Exec executes command from commands map. It returns the result of the command
// execution or an error if the command is not found.
//
// Parameters:
// - command: The name of the command to execute.
// - processIn: The type of input processing to use for the command.
// - data: The data to pass to the command handler.
//
// Returns:
// - []byte: The result of the command execution.
// - error: An error if the command is not found.
func (c *Commands) Exec(command string, processIn ProcessIn, data any) (
	[]byte, error) {

	// Get the command from the commands map by name.
	cmd, ok := c.Get(command)

	// If the command is found and has a handler, execute the handler.
	if ok && cmd.Handler != nil {
		return cmd.Handler(cmd, processIn, data)
	}

	// If the command is not found, return an error.
	return nil, fmt.Errorf("command '%s' not found", command)
}

// ForEach calls the given function for each added command.
//
// The function passed as an argument should accept two parameters:
// - command (string): the name of the command
// - cmd (*CommandData): the data of the command
//
// Example usage:
//
//	// Prints the name of each added command
//	commands.ForEach(func(command string, cmd *CommandData) {
//	    fmt.Println(command)
//	})
//
// Deprecated: This method is deprecated and will be removed in the future.
// Use Iter instead.
//
//go:deprecated
func (c *Commands) ForEach(f func(command string, cmd *CommandData)) {
	// Lock the commands map for reading to prevent concurrent modifications
	c.RLock()
	defer c.RUnlock()

	// Iterate over the commands map and call the given function for each command
	for command, cmd := range c.m {
		f(command, cmd)
	}
}

// Iter returns an iterator for the commands map.
//
// This function is safe for concurrent read access. But all write Commands
// methods will lock during iteration.
func (c *Commands) Iter() iter.Seq2[string, *CommandData] {
	return func(yield func(string, *CommandData) bool) {
		c.RLock()
		defer c.RUnlock()

		for command, cmd := range c.m {
			if !yield(command, cmd) {
				return
			}
		}
	}
}

// HabdleCommands is a function that adds handlers to the commands added to the
// Commands struct. It takes two parameters:
//   - processIn: a ProcessIn variable that specifies the input processing types
//     for the commands to be handled.
//   - h: a function that specifies how to handle the added commands. The function
//     takes two parameters:
//   - command: a string that represents the name of the command.
//   - params: a string that represents the parameters of the command.
//
// The function iterates over the commands map using the ForEach method and
// checks if the command's ProcessIn field has any bitwise AND operation with
// the processIn parameter and if the command's Handler field is not nil. If both
// conditions are true, the h function is called with the command's name,
// parameters, and handler.
func (c *Commands) HabdleCommands(processIn ProcessIn,
	handler func(command, params string)) {

	for command, cmd := range c.Iter() {
		if cmd.ProcessIn&processIn != 0 && cmd.Handler != nil {
			handler(command, cmd.Params)
		}
	}
}

// ParseCommand parses the given input command data.
//
// It returns the CommandData
// associated with the command name, the name of the command, a map of
// variables, the command data, and an error if the command is not found.
//
// The input data is split by / on two parts: name and parameters. The name
// is used to look up the command in the Commands map. If the command is not
// found, an error is returned. The parameters are split by / on parts with
// length of command parameters + 1. The last part is the command data. The
// command parameters and its values are used to create a map of variables.
//
// If the command data is present, it is returned as is or nil if command data
// is not present.
func (c *Commands) ParseCommand(inData []byte) (cmd *CommandData, name string,
	vars map[string]string, data []byte, err error) {

	// Split data (parameters) by / on two parts: name and parameters
	v := bytes.SplitN(inData, []byte("/"), 2)
	name = string(v[0])
	var parameters []byte
	if len(v) > 1 {
		parameters = v[1]
	}

	// Look up the command in the Commands map
	cmd, ok := c.Get(name)
	if !ok {
		// Return an error if the command is not found
		err = fmt.Errorf("command '%s' not found", name)
		return
	}

	// Create a map of variables
	vars = make(map[string]string)

	// If there is no arameters, return cmd, name and empty var
	// and parts
	if len(parameters) == 0 {
		return
	}

	// Command parameters
	params := cmd.ParamsSlice()

	// Split input parameters by / on parts with lenght of params + 1
	parts := bytes.SplitN(parameters, []byte("/"), len(params)+1)

	// Create a map of variables from command parameters and its values
	for i, param := range params {

		// If name of command parameter is empty than skip it
		if len(param) == 0 {
			continue
		}

		// Get the value of the parameter by index of the parameter and assign
		// it to vars map
		var v string
		if len(parts) > i {
			v = string(parts[i])
		}
		vars[param] = v
	}

	// The last part of the parts of input parameters is the command data
	if len(parts) > len(params) {
		data = parts[len(parts)-1]
	}

	return
}

// Print prints list of added commands.
func (c *Commands) Print(msgs ...string) {
	for command, cmd := range c.Iter() {
		fmt.Printf("---\n"+
			"%s\n"+
			"%s\n"+
			"parameters: %s\n"+
			"processing in: %s\n"+
			"\n",
			command, cmd.Descr, cmd.Params, cmd.ProcessIn)
	}
}
