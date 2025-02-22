package command

import (
	"fmt"
	"testing"
)

func TestParseCommand(t *testing.T) {

	c := New()
	c.Add("test", "test", HTTP, "{param1}/{param2}/{param3}", "", "", "",
		func(cmd *CommandData, processIn ProcessIn, data any) ([]byte, error) {
			return []byte("test"), nil
		},
	)

	tst := func(message []byte) {
		_, name, vars, data, _ := c.ParseCommand(message)
		fmt.Println(string(message), " -> ", name, vars)

		cmd, ok := c.Get(name)
		if !ok {
			return
		}

		params := cmd.ParamsSlice()
		for _, v := range params {
			fmt.Println(v, vars[v])
		}

		if len(data) > 0 {
			fmt.Println("data:", string(data))
		}

		fmt.Println()
	}

	tst([]byte("test/value1"))

	tst([]byte("test/value1/value2/value3"))

	tst([]byte("test/value1/{\"value2/subvalue\"}/value3"))

	tst([]byte("test/value1/value2/{\"json string with slashes/subvalue\"}"))

	// A value with slashes (or some binary data) will be processed successfully
	// only in the last parameter - data
	tst([]byte("test/value1/value2/value3/{\"json string with slashes/subvalue\"}"))
}
