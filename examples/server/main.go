// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This sample application demonstrates how to use the Command package with
// HTTP server.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/kirill-scherba/command/v2"
)

const (
	appShort   = "server"
	appVersion = "0.0.1"
	appPort    = "8084"

	apiprefix = "/api/v1/"
)

// Application parameters type.
type Parameters struct {
	addr string // HTTP address
	port string // HTTP port
}

// Application parameters object.
var params Parameters

func main() {

	// Application Logo
	fmt.Printf("Command package example server application ver. %s\n", appVersion)

	// Get HTTP port from environment variable
	// params.port = os.Getenv("PORT")
	if params.port == "" {
		params.port = appPort
	}
	fmt.Println("HTTP port:", params.port)

	// Parse parameters
	flag.StringVar(&params.addr, "addr", ":"+params.port, "http server local address")
	flag.Parse()

	// Create command object
	c := command.New()

	// Add commands
	commands(c)

	// Start HTTP server
	serve(c)
}

// Server commands
func commands(c *command.Commands) {

	// Add 'hello' commands
	c.Add("hello", "say hello", command.HTTP|command.WS, "{name}", "", "", "",
		func(cmd *command.CommandData, processIn command.ProcessIn, data any) (
			io.Reader, error) {

			log.Println("executing command 1: hello", data)

			vars, err := c.Vars(data)
			if err != nil {
				return nil, err
			}

			log.Println("executing command 2: hello", data, vars)

			return strings.NewReader(fmt.Sprintf("Hello %s!", vars["name"])), nil
		},
	)

	// Add 'version' commands
	c.Add("version", "get application version", command.HTTP|command.WS, "", "", "", "",
		func(cmd *command.CommandData, processIn command.ProcessIn, data any) (
			io.Reader, error) {

			return strings.NewReader(appVersion), nil
		},
	)

	// Add commands list
	c.AddCommandsList(command.HTTP)
}
