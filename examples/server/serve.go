// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Create and start a HTTP server and handle commands.

package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kirill-scherba/command/v2"
)

// HttpRequest contains gorilla mux variables and HTTP request.
type HttpRequest struct {
	*http.Request
	Vars map[string]string
}

func (r *HttpRequest) GetVars() map[string]string {
	return r.Vars
}

func (r *HttpRequest) GetData() []byte {
	return nil
}

func serve(c *command.Commands) {
	// Create a mux for routing incoming requests
	m := mux.NewRouter()

	// Commands HTTP handlers
	c.HabdleCommands(command.HTTP, func(name, params string) {

		// Handler path
		path := apiprefix + name
		path = strings.TrimRight(path, "/")
		if len(params) > 0 {
			path += "/" + params
		}

		// Add HTTP handler
		m.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

			// Handlers request contains gorilla mux variables and HTTP request
			request := &HttpRequest{r, mux.Vars(r)}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")

			// Execute command
			data, err := c.Exec(name, command.HTTP, request)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Write response
			w.Write([]byte(data))
		})

	})

	// WebSocket handler
	serveWs(m, c)

	// Local file system files
	frontendFS := http.FileServer(http.FS(getFrontendDistFs()))
	m.PathPrefix("/").Handler(frontendFS)

	// Start HTTP server
	log.Printf("start listening for HTTP requests on %s, http://localhost:%s",
		params.addr, params.port)
	log.Fatalln(http.ListenAndServe(params.addr, m))
}
