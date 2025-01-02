// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Handle and process HTTP websocket commands.

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kirill-scherba/command/v2"
)

// WsRequest contains gorilla websocket connection and variables map.
type WsRequest struct {
	*websocket.Conn
	Vars map[string]string
}

func (r *WsRequest) GetVars() map[string]string {
	return r.Vars
}

func (r *WsRequest) GetData() []byte {
	return nil
}

// ServeWs handles and processes HTTP websocket commands.
type ServeWs struct {
	c    *command.Commands
	conn *websocket.Conn
}

// serveWs start a HTTP websocket handler.
func serveWs(m *mux.Router, c *command.Commands) {
	m.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		// Upgrade HTTP connection to WebSocket
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade connection:", err)
			return
		}

		// Handle WebSocket connection
		go (&ServeWs{c, conn}).handleConnection(conn)
	})
}

// handleConnection handles the connection with a client.
//
// It takes a pointer to a websocket.Conn as a parameter.
func (s *ServeWs) handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("failed to read message from client:", err)
			break
		}

		// Process message
		s.processMessage(conn, message)
	}
}

func (s *ServeWs) processMessage(conn *websocket.Conn, message []byte) {
	// Print message to console
	log.Println("received message:", string(message))

	// Parse message
	name, vars := s.c.ParseCommand(message)

	// Execute command
	log.Println("executing command:", name, vars)
	res, err := s.c.Exec(name, command.WS, &WsRequest{Conn: conn, Vars: vars})
	if err != nil {
		log.Println("failed to execute command:", err)
		res = []byte(err.Error())
	}

	// Write answer
	s.conn.WriteMessage(websocket.TextMessage, res)
}
