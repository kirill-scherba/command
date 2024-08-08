// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ProcessIn module of Command processing golang package.

package command

import "strings"

const (
	HTTP   ProcessIn = 1 << iota // HTTP request
	TRU                          // TRU request
	WebRTC                       // WebRTC request
	Teonet                       // Teonet request
	WS                           // Websocket request
	All    = HTTP | TRU | WebRTC | Teonet | WS
)

// ProcessIn represents the source of a command.
type ProcessIn byte

// String returns a string representation of the ProcessIn.
//
// The string representation includes the names of the sources separated by commas.
// If the source is unknown, it is omitted from the result.
// The result is lowercased and trimmed of trailing commas.
func (pi ProcessIn) String() string {
	// Create a strings.Builder to efficiently build the result string.
	var sb strings.Builder

	// Check each bit of the ProcessIn byte and append the corresponding source name
	// to the strings.Builder if the bit is set.

	// HTTP source
	if pi&HTTP != 0 {
		sb.WriteString("HTTP, ")
	}

	// TRU source
	if pi&TRU != 0 {
		sb.WriteString("TRU, ")
	}

	// WebRTC source
	if pi&WebRTC != 0 {
		sb.WriteString("WebRTC, ")
	}

	// Teonet source
	if pi&Teonet != 0 {
		sb.WriteString("Teonet, ")
	}

	// Websocket source
	if pi&WS != 0 {
		sb.WriteString("Websocket, ")
	}

	// Get the result string from the strings.Builder.
	result := sb.String()

	// Trim the trailing comma and space from the result string.
	result = strings.TrimRight(result, ", ")

	// Convert the result string to lowercase.
	result = strings.ToLower(result)

	// Return the result string.
	return result
}
