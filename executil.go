package executil

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"

	"gitlab.home.mikenewswanger.com/golang/filesystem"
)

// Command describes the execution instructions for the command to be run
type Command struct {
	Name              string
	WorkingDirectory  string
	Executable        string
	Arguments         []string
	Debug             bool
	Verbosity         uint8
	ContinueOnFailure bool
	cmd               *exec.Cmd
}

// RunWithRealtimeOutput runs the given command and if verbosity is set, returns real-time output to stdout and stderr
func (c Command) RunWithRealtimeOutput() {
	color.Green(c.Name)
	c.setExecutionEnvironment()

	var stderrBuffer bytes.Buffer
	c.cmd.Stderr = &stderrBuffer

	if c.Verbosity > 0 {
		c.cmd.Stdout = os.Stdout
		c.cmd.Stderr = os.Stderr
	}

	if c.run() {
		color.Green("Success")
		println()
	} else {
		if !c.ContinueOnFailure {
			if c.Verbosity == 0 {
				color.Red(stderrBuffer.String())
			} else {
				color.Red("Failed to execute command")
			}
			os.Exit(2)
		}
	}
}

// RunWithCombinedOutput executes the given command then returns its combined stdout and stderr to the caller
func (c Command) RunWithCombinedOutput() []byte {
	c.setExecutionEnvironment()
	output, _ := c.cmd.CombinedOutput()
	return output
}

// RunWithOutput executes the given command then returns its stdout to the caller
func (c Command) RunWithOutput() ([]byte, bool) {
	c.setExecutionEnvironment()
	output, err := c.cmd.Output()
	return output, err == nil
}

func (c *Command) setExecutionEnvironment() {
	c.cmd = exec.Command(c.Executable, c.Arguments...)
	if c.WorkingDirectory != "" {
		c.cmd.Dir = filesystem.BuildAbsolutePathFromHome(c.WorkingDirectory)
	}
	if c.Debug {
		color.Yellow(c.Executable + " " + strings.Join(c.Arguments, " "))
	}
}

func (c *Command) run() bool {
	if err := c.cmd.Start(); err != nil {
		color.Red("Could not start process")
		panic(err)
	}

	err := c.cmd.Wait()
	return err == nil
}
