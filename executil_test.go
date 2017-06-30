package executil

import (
	"testing"
)

func TestMe(t *testing.T) {
	SetVerbosity(4)
	var c = Command{
		Name:       "Test function output",
		Executable: "echo",
		Arguments: []string{
			"It works!",
		},
	}
	if err := c.Run(); err != nil {
		t.Error("Command failed to execute")
	}
	println(c.GetStdout())
	println(c.GetStderr())
}
