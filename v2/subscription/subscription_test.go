package subscription

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DataChannel struct {
	ConnectionChannel
}

func TestAddNewActionToEmptySubscription(t *testing.T) {
	sub := New()
	con := DataChannel{}
	cmd := "test-command"
	handler := func(command string, data any) ([]byte, error) { return nil, nil }

	sub.SubscribeCmd(con, cmd, nil, handler)

	assert.NotNil(t, sub.SubscribersMap[con])
	assert.NotNil(t, sub.ActionsMap[ActionCommand(cmd)])
	assert.Equal(t, cmd, sub.SubscribersMap[con][ActionCommand(cmd)].Command)
}

func TestAddNewActionToExistingSubscription(t *testing.T) {
	sub := New()
	con := DataChannel{}
	cmd1 := "test-command-1"
	cmd2 := "test-command-2"
	handler1 := func(command string, data any) ([]byte, error) { return nil, nil }
	handler2 := func(command string, data any) ([]byte, error) { return nil, nil }

	sub.SubscribeCmd(con, cmd1, nil, handler1)
	sub.SubscribeCmd(con, cmd2, nil, handler2)

	assert.NotNil(t, sub.SubscribersMap[con])
	assert.NotNil(t, sub.ActionsMap[ActionCommand(cmd1)])
	assert.NotNil(t, sub.ActionsMap[ActionCommand(cmd2)])
	assert.Equal(t, cmd1, sub.SubscribersMap[con][ActionCommand(cmd1)].Command)
	assert.Equal(t, cmd2, sub.SubscribersMap[con][ActionCommand(cmd2)].Command)
}

func TestAddSameActionMultipleTimes(t *testing.T) {
	sub := New()
	con := DataChannel{}
	cmd := "test-command"
	handler := func(command string, data any) ([]byte, error) { return nil, nil }

	sub.SubscribeCmd(con, cmd, nil, handler)
	sub.SubscribeCmd(con, cmd, nil, handler)

	assert.NotNil(t, sub.SubscribersMap[con])
	assert.NotNil(t, sub.ActionsMap[ActionCommand(cmd)])
	assert.Equal(t, cmd, sub.SubscribersMap[con][ActionCommand(cmd)].Command)
}

func TestAddDifferentActionsToSameSubscription(t *testing.T) {
	sub := New()
	dc1 := DataChannel{}
	dc2 := DataChannel{}
	cmd1 := "test-command-1"
	cmd2 := "test-command-2"
	handler1 := func(command string, data any) ([]byte, error) { return nil, nil }
	handler2 := func(command string, data any) ([]byte, error) { return nil, nil }

	sub.SubscribeCmd(dc1, cmd1, nil, handler1)
	sub.SubscribeCmd(dc2, cmd2, nil, handler2)

	assert.NotNil(t, sub.SubscribersMap[dc1])
	assert.NotNil(t, sub.SubscribersMap[dc2])
	assert.NotNil(t, sub.ActionsMap[ActionCommand(cmd1)])
	assert.NotNil(t, sub.ActionsMap[ActionCommand(cmd2)])
	assert.Equal(t, cmd1, sub.SubscribersMap[dc1][ActionCommand(cmd1)].Command)
	assert.Equal(t, cmd2, sub.SubscribersMap[dc2][ActionCommand(cmd2)].Command)
}

func TestConcurrentAccess(t *testing.T) {
	sub := New()
	con := DataChannel{}
	cmd := "test-command"
	handler := func(command string, data any) ([]byte, error) { return nil, nil }

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			sub.SubscribeCmd(con, cmd, nil, handler)
			wg.Done()
		}()
	}
	wg.Wait()

	assert.NotNil(t, sub.SubscribersMap[con])
	assert.NotNil(t, sub.ActionsMap[ActionCommand(cmd)])
	assert.Equal(t, cmd, sub.SubscribersMap[con][ActionCommand(cmd)].Command)
}

type Con struct{}

func (Con) GetUser() interface{}     { return nil }
func (Con) SetUser(user interface{}) {}
func (Con) Send(data []byte) error   { fmt.Printf("send %s done\n", string(data)); return nil }

// Test Exec
func TestExec(t *testing.T) {
	sub := New()

	var con Con
	cmd := "test-command"

	handler := func(command string, data any) ([]byte, error) {
		d := data.(string)
		fmt.Printf("process command: %s, data: '%s'\n", command, d)
		return []byte(d), nil
	}

	sub.SubscribeCmd(con, cmd, "some data to process command", handler)
	sub.ExecCmd(cmd)
}

func TestExistsDcCmd(t *testing.T) {
	sub := New()
	var con Con
	cmd := "test-command"

	tests := []struct {
		name     string
		con      Con
		command  string
		expected bool
	}{
		{
			name:     "command exists for data channel",
			con:      con,
			command:  cmd,
			expected: true,
		},
		{
			name:     "command does not exist for data channel",
			con:      con,
			command:  "non-existent-command",
			expected: false,
		},
		// {
		// 	name:     "data channel does not exist in SubscribersMap",
		// 	con:       con{},
		// 	command:  cmd,
		// 	expected: false,
		// },
		// {
		// 	name:     "empty command",
		// 	con:       con,
		// 	command:  "",
		// 	expected: false,
		// },
		// {
		// 	name:     "nil Subscription",
		// 	con:       con,
		// 	command:  cmd,
		// 	expected: false,
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "command exists for data channel" {
				sub.SubscribeCmd(con, cmd, nil, func(command string, data any) ([]byte, error) { return nil, nil })
			}

			actual := sub.ExistsConCmd(test.con, test.command)
			if actual != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}
