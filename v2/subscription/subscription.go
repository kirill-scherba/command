// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package subscription process commands subscriber.
package subscription

import (
	"encoding/json"
	"iter"
	"log"
	"sync"
)

// Subscription type stores subscribers.
type Subscription struct {

	// Mutex for concurrent access to subscribers object
	mut *sync.RWMutex

	// Subscribers map to store commands by connection channel and action
	// (command name and action func)
	Subscribers
}

// Subscribers type to store commands by connection channel and action
// (command name and action func).
type Subscribers struct {
	SubscribersMap
	ActionsMap
}

// SubscribersMap map to store actions by connection channel and action.
type SubscribersMap map[ConnectionChannel]map[ActionCommand]*SubscribersAction

// ActionsMap map to store actions by command name and connection channel.
type ActionsMap map[ActionCommand]map[ConnectionChannel]*SubscribersAction

// ActionCommand represents command name.
type ActionCommand string

// SubscribersAction represents action for subscribers.
// It contains command name and handler function.
type SubscribersAction struct {
	// Request command or Action name
	Command string

	// Request data
	Data any

	// Handler function to process command
	Handler func(command string, data any) ([]byte, error)
}

// ConnectionChannel represents connection channel interface.
type ConnectionChannel interface {
	GetUser() any
	SetUser(user any)
	Send(data []byte) error
}

// Teogw packet data
type TeogwData struct {
	ID      uint32 `json:"id"`
	Address string `json:"address"`
	Command string `json:"command"`
	Data    []byte `json:"data"`
	Err     string `json:"err"`
}

// New creates new Subscription object.
func New() *Subscription {
	return &Subscription{
		mut: new(sync.RWMutex),
		Subscribers: Subscribers{
			SubscribersMap: make(SubscribersMap),
			ActionsMap:     make(ActionsMap),
		},
	}
}

// SubscribeCmd adds subscribers command to Subscription.
func (s *Subscription) SubscribeCmd(con ConnectionChannel, command string, data any,
	handler func(command string, data any) ([]byte, error)) {

	s.mut.Lock()
	defer s.mut.Unlock()

	log.Printf("subscribe command: %s", command)

	// Command
	cmd := ActionCommand(command)

	// Action
	action := &SubscribersAction{Command: command, Data: data, Handler: handler}

	// Add new action to subscribers map
	if s.SubscribersMap[con] == nil {
		s.SubscribersMap[con] = map[ActionCommand]*SubscribersAction{}
	}
	s.SubscribersMap[con][cmd] = action

	// Add new action to actions map
	if s.ActionsMap[cmd] == nil {
		s.ActionsMap[cmd] = map[ConnectionChannel]*SubscribersAction{}
	}
	s.ActionsMap[cmd][con] = action
}

// DelCon deletes all subscriptions by connection channel.
func (s *Subscription) DelCon(con ConnectionChannel) {
	s.mut.Lock()
	defer s.mut.Unlock()

	log.Println("unsubscribe connection")

	// Delete connection channel from subscribers map
	delete(s.SubscribersMap, con)

	// Delete action from actions map
	for k, v := range s.ActionsMap {
		if _, ok := v[con]; ok {
			delete(v, con)
			if len(v) == 0 {
				delete(s.ActionsMap, k)
			}
		}
	}
}

// DelConCmd deletes subscriptions by connection channel and command name.
func (s *Subscription) DelConCmd(con ConnectionChannel, command string) {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Delete action from actions map
	if _, ok := s.ActionsMap[ActionCommand(command)][con]; ok {
		delete(s.ActionsMap[ActionCommand(command)], con)
		if len(s.ActionsMap[ActionCommand(command)]) == 0 {
			delete(s.ActionsMap, ActionCommand(command))
		}
	}

	// Delete action from subscribers map
	if _, ok := s.SubscribersMap[con][ActionCommand(command)]; ok {
		delete(s.SubscribersMap[con], ActionCommand(command))
		if len(s.SubscribersMap[con]) == 0 {
			delete(s.SubscribersMap, con)
		}
	}
}

// ExecCmd executes command for all connection channels and send response to
// subscribers.
func (s *Subscription) ExecCmd(command string) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	// Get actions
	actions, ok := s.ActionsMap[ActionCommand(command)]
	if !ok || len(actions) == 0 {
		return
	}

	// Send action result to all subscribers connection channels
	var wg sync.WaitGroup
	for con, a := range actions {
		wg.Add(1)
		go func(con ConnectionChannel, a *SubscribersAction) {
			defer wg.Done()

			// Execute action
			data, err := a.Handler(a.Command, a.Data)
			log.Printf("process command: %s, data len: '%d'\n", a.Command,
				len(data))

			// Marshal data
			d, _ := json.Marshal(TeogwData{
				Command: command,
				Data:    data,
				Err: func(err error) (errStr string) {
					if err != nil {
						errStr = err.Error()
					}
					return
				}(err),
			})

			// Send command to connection channel
			con.Send(d)
		}(con, a)
	}
	wg.Wait()
}

// ExecConCmd executes command for all subscriptions by selected connection
// channel.
func (s *Subscription) ExecConCmd(con ConnectionChannel, command string) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	// Get action
	action, ok := s.SubscribersMap[con][ActionCommand(command)]
	if !ok {
		return
	}

	// Execute action
	data, err := action.Handler(action.Command, action.Data)
	log.Printf("process command: %s, data len: %d\n", action.Command, len(data))

	// Marshal data
	d, _ := json.Marshal(TeogwData{
		Command: command,
		Data:    data,
		Err: func(err error) (errStr string) {
			if err != nil {
				errStr = err.Error()
			}
			return
		}(err),
	})

	// Send command to webrtc connection channel
	con.Send(d)
}

// ExistsConCmd checks if command exists for selected connection channel.
func (s *Subscription) ExistsConCmd(con ConnectionChannel, command string) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	_, ok := s.SubscribersMap[con][ActionCommand(command)]

	return ok
}

// UpdateConCmd updates command data for selected connection channel.
func (s *Subscription) UpdateConCmd(con ConnectionChannel, command string,
	data any) {

	s.mut.RLock()
	defer s.mut.RUnlock()

	action, ok := s.SubscribersMap[con][ActionCommand(command)]
	if ok {
		action.Data = data
	}
}

// DataConCmd get command data for selected connection channel.
func (s *Subscription) DataConCmd(con ConnectionChannel, command string) any {
	s.mut.RLock()
	defer s.mut.RUnlock()

	action, ok := s.SubscribersMap[con][ActionCommand(command)]
	if ok {
		return action.Data
	}
	return nil
}

// Iter returns an iterator for all subscribed commands.
func (s *Subscription) Iter() iter.Seq[string] {
	return func(yield func(string) bool) {
		s.mut.RLock()
		defer s.mut.RUnlock()

		for command := range s.ActionsMap {
			if !yield(string(command)) {
				return
			}
		}
	}
}
