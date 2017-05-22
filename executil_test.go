package executil

import (
	"testing"
)

func TestMe(t *testing.T) {
	var c = Command{
		Name:       "Test function output",
		Executable: "docker",
		Verbosity:  4,
	}
	if err := c.Run(); err != nil {
		t.Error("Command failed to execute")
	}
	println(c.GetStdout())
	println(c.GetStderr())
}