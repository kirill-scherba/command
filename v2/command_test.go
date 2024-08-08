package command

import (
	"fmt"
	"testing"
)

func TestParseCommand(t *testing.T) {

	c := New()
	c.Add("test", "test", HTTP, "{param1}/{param2}/{param3}",
		func(cmd *CommandData, processIn ProcessIn, data any) ([]byte, error) {
			return []byte("test"), nil
		},
	)

	tst := func(message []byte) {
		name, vars := c.ParseCommand(message)
		fmt.Println(string(message), " -> ", name, vars)

		cmd, ok := c.Get(name)
		if !ok {
			return
		}
		params := cmd.ParamsSlice()
		for _, v := range params {
			fmt.Println(v, vars[v])
		}
		fmt.Println()
	}

	tst([]byte("test/value1"))

	tst([]byte("test/value1/value2/value3"))

	tst([]byte("test/value1/{\"value2/subvalue\"}/value3"))

	// Value with slashes processed successfully in last parameter only
	tst([]byte("test/value1/value2/{\"json string with slashes/subvalue\"}"))
}
