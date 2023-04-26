package command

import (
	"fmt"
	"testing"
)

func TestCommand(t *testing.T) {

	// Create command object
	com := New()

	// Add some commands
	com.Add([]*CommandData{
		// Help command
		// {
		// 	name:  "help",
		// 	usage: "show this help message",
		// 	cmd: func(params ...string) (res []byte, err error) {
		// 		res = []byte("\nUsage of commands:\n" + com.Usage())
		// 		return
		// 	},
		// },

		// Get some text commands
		{
			Name:  "get",
			Usage: "return some text",
			Cmd: func(params ...string) (res []byte, err error) {
				res = []byte("this is some text")
				return
			},
		},

		// Get some text wit parameter commands
		{
			Name:   "Getparam",
			Usage:  "return some text",
			Params: []ParamData{{"<param1>", "simple parameter"}},
			Cmd: func(params ...string) (res []byte, err error) {
				res = []byte(fmt.Sprintf("this is some text with %s", params[0]))
				return
			},
		},
	}...)

	showResult := func(result []byte, err error) {
		if err != nil {
			// t.Log(err)
		} else {
			t.Log(string(result))
		}
	}

	// Execute commands

	// Help command
	t.Log("'Help' command")
	res, err := com.Exec([]byte("help"))
	showResult(res, err)
	if err != nil {
		t.Error(err)
		return
	}

	// Get command
	t.Log("'Get' command")
	res, err = com.Exec([]byte("get"))
	showResult(res, err)
	if err != nil {
		t.Error(err)
		return
	}

	// Get command (uppercase)
	t.Log("'Get' command (uppercase)")
	res, err = com.Exec([]byte("Get"))
	showResult(res, err)
	if err != nil {
		t.Error(err)
		return
	}

	// Getparam command
	t.Log("'Getparam' command")
	res, err = com.Exec([]byte("Getparam parameter"))
	showResult(res, err)
	if err != nil {
		t.Error(err)
		return
	}

	// Wrong (not exsisted) command
	t.Log("'Wrong' command")
	res, err = com.Exec([]byte("Wrong"))
	showResult(res, err)
	if err == nil {
		t.Error(err)
		return
	}
}
