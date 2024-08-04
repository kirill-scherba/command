// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ProcessIn module of Command processing golang package.

package command

import "strings"

const (
	HTTP ProcessIn = 1 << iota
	TRU
	WebRTC
)

type ProcessIn byte

func (p ProcessIn) String() string {

	var s string
	if p&HTTP != 0 {
		s += "HTTP, "
	}
	if p&TRU != 0 {
		s += "TRU, "
	}
	if p&WebRTC != 0 {
		s += "WebRTC, "
	}
	s = strings.Trim(s, ", ")
	s = strings.ToLower(s)
	return s
}
