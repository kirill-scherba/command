// Copyright 2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command go package organithe Client/Server command processing.
package command

import (
	"fmt"
	"strings"
	"sync"
)

type Command struct {
	m        CommandMap
	mut      *sync.RWMutex
	commands []*CommandData
}
type CommandMap map[string]*CommandData
type CommandData struct {
	Name   string
	Usage  string
	Params []ParamData
	Cmd    CommandFunc
}
type ParamData struct {
	Name  string
	Usage string
}
type CommandFunc func(params ...string) ([]byte, error)

// New creates new Command object
func New() (c *Command) {
	c = new(Command)
	c.m = make(CommandMap)
	c.mut = new(sync.RWMutex)
	c.addHelp()
	return
}

// addHelp add predefined help command
func (c *Command) addHelp() {
	c.Add(&CommandData{
		Name:  "help",
		Usage: "show this help message",
		Cmd: func(params ...string) (res []byte, err error) {
			res = []byte(fmt.Sprintf("\nUsage of commands:\n%s", c))
			return
		},
	})
}

// String returns string with command Usage definitions
func (c Command) String() (usage string) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	const (
		INDENT   = "  "
		DASH     = " - "
		NEW_LINE = "\n"
	)

	for i := range c.commands {
		if i > 0 {
			usage += NEW_LINE
		}
		usage += INDENT + c.commands[i].Name
		for j := range c.commands[i].Params {
			usage += " " + c.commands[i].Params[j].Name
		}
		usage += DASH + c.commands[i].Usage
		for j := range c.commands[i].Params {
			usage += NEW_LINE +
				INDENT + INDENT + c.commands[i].Params[j].Name +
				DASH + c.commands[i].Params[j].Usage
		}
	}

	return
}

// Add adds command
func (c *Command) Add(cmds ...*CommandData) (err error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, cmd := range cmds {
		if cmd == nil {
			err = fmt.Errorf("command pointer is nil")
			return
		}
		if len(cmd.Name) == 0 {
			err = fmt.Errorf("command name is empty")
			return
		}
		cmd.Name = strings.ToLower(cmd.Name)
		c.findAndReplace(cmd)
		c.m[cmd.Name] = cmd
	}

	return
}

// findAndReplace find command by name in commands array and replace if exists
// or add to array if not exists
func (c *Command) findAndReplace(cmd *CommandData) {
	for i := range c.commands {
		if c.commands[i].Name == cmd.Name {
			c.commands[i] = cmd
			return
		}
	}
	c.commands = append(c.commands, cmd)
}

// Get gets command by name
func (c Command) Get(name string) (commandData *CommandData, ok bool) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	commandData, ok = c.m[name]
	return
}

// Exec executes command
func (c Command) Exec(cmd []byte) (result []byte, err error) {

	params := strings.Fields(string(cmd))

	// Check name set
	if len(params) == 0 {
		err = fmt.Errorf("name should be set")
		return
	}
	name := strings.ToLower(params[0])
	params = params[1:]

	// Check name
	commandData, ok := c.Get(name)
	if !ok {
		err = fmt.Errorf("command %s not found", name)
		return
	}

	// Check command function defined
	if commandData.Cmd == nil {
		err = fmt.Errorf("command %s is not defined", name)
		return
	}

	// Execute command
	result, err = commandData.Cmd(params...)
	return
}
