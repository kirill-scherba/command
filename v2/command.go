// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command processing golang package.
package command

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrIncorrectInputData is an error returned when the input data provided is
// incorrect.
var ErrIncorrectInputData = fmt.Errorf("inclorrect input data")

// Commands is a struct that contains a map of command data and a read-write
// mutex for synchronizing access to the map.
type Commands struct {
	m map[string]*CommandData
	*sync.RWMutex
}

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
		// and append it to the params slice.
		params = append(params, strings.Trim(param, "{}"))
	}

	// Return the params slice.
	return params
}

// CommandHandler is a function that handles a command.
type CommandHandler func(cmd *CommandData, processIn ProcessIn, data any) ([]byte, error)

// RequestInterface is commont type of Requesr interface.
type RequestInterface interface {
	// GetVars returns map of request variables.
	GetVars() map[string]string

	// GetData returns request data.
	GetData() []byte

	// Setd date to responce. Used in HTTP request and set custom date to HTTP
	// writer.
	SetDate(date time.Time)
}

// ParseParams parses the input data command parameters.
// It attempts to assert the input data to the specified type T.
// If the assertion is successful, it returns the parsed request and a nil error.
// If the assertion fails, it returns the zero value of T and an error indicating that the input data was incorrect.
//
// Parameters:
//   - command: a pointer to a CommandData struct representing the command being executed
//   - indata: the input data to be parsed
//
// Returns:
//   - T: the parsed request of type T
//   - error: an error indicating if the input data was incorrect
func ParseParams[T any]( /* command *CommandData,  */ indata any) (T, error) {
	// Initialize the error variable
	var err error

	// Attempt to assert the input data to the specified type T
	request, ok := indata.(T)

	// If the assertion fails, set the error variable and log the error
	if !ok {
		err = ErrIncorrectInputData
		// log.Printf("%s parse params error: %s", command.Cmd, err)
	}

	// Return the parsed request and the error
	return request, err
}

// New create and initialize Commands object.
func New() *Commands {
	c := &Commands{}
	c.Init()
	return c
}

// Init initialize Commands object and add default commands.
func (c *Commands) Init() {
	c.m = make(map[string]*CommandData)
	c.RWMutex = new(sync.RWMutex)
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
	c.m[command] = &CommandData{
		command, processIn, params, returnDescr, descr, request, response, handler,
	}
	c.Unlock()
	return c
}

// Get returns CommandData from commands map by name.
func (c *Commands) Get(name string) (cmd *CommandData, ok bool) {
	c.RLock()
	cmd, ok = c.m[name]
	c.RUnlock()
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
func (c *Commands) ForEach(f func(command string, cmd *CommandData)) {
	// Lock the commands map for reading to prevent concurrent modifications
	c.RLock()
	defer c.RUnlock()

	// Iterate over the commands map and call the given function for each command
	for command, cmd := range c.m {
		f(command, cmd)
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
	h func(command, params string)) {

	c.ForEach(func(command string, cmd *CommandData) {
		if cmd.ProcessIn&processIn != 0 && cmd.Handler != nil {
			h(command, cmd.Params)
		}
	})
}

// ParseCommand parses the command data.
//
// It takes a byte slice 'data' representing the command data and returns the
// command name and a map of variables.
//
// The function splits the 'data' by the '/' character and extracts the command
// name. It then retrieves the command data for the command name from the
// Commands struct. If the command is not found, it returns an empty name and
// an empty map of variables.
//
// The function then splits the remaining command parameters by the '/'
// character and creates a map of variables. Each variable is a key-value pair,
// where the key is the parameter name and the value is the parameter value.
//
// ***A value with slashes is processed successfully only in the last parameter
// in 'data'.
func (c *Commands) ParseCommand(data []byte) (name string, vars map[string]string) {

	// Initialize a map to store the command variables
	vars = make(map[string]string)

	// Split the command data by '/' character
	v := bytes.SplitN(data, []byte("/"), 2)

	// Set the command name as the first part of the split data
	name = string(v[0])

	// If there is no second part, return empty name and empty variables
	if len(v) < 2 {
		return
	}

	// Get the command parameters
	cmdParams := v[1]

	// Get the command data for the command name from the Commands struct
	cmd, ok := c.Get(name)
	if !ok {
		// If the command is not found, return empty name and empty variables
		return
	}

	// Get the command parameters as a slice
	params := cmd.ParamsSlice()

	// If there are command parameters, create a map of variables
	if len(cmdParams) > 0 {

		// Split the command parameters by '/' character and create a map of
		// variables
		for i, v := range bytes.SplitN(cmdParams, []byte("/"), len(params)) {
			vars[params[i]] = string(v)
		}
	}

	// Return the command name and the map of variables
	return
}
