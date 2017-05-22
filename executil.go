package executil

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/filesystem"
)

// Command describes the execution instructions for the command to be run
type Command struct {
	Name             string
	Arguments        []string
	Executable       string
	Logger           *logrus.Logger
	Verbosity        uint8
	WorkingDirectory string
	cmd              *exec.Cmd
	loggerFields     logrus.Fields
	stdout           string
	stderr           string
	validationErrors []string
	waitGroup        *sync.WaitGroup
}

func (c *Command) initialize() {
	if c.Logger == nil {
		c.Logger = logrus.New()

		switch c.Verbosity {
		case 0:
			c.Logger.Level = logrus.ErrorLevel
			break
		case 1:
			c.Logger.Level = logrus.WarnLevel
			break
		case 2:
			fallthrough
		case 3:
			c.Logger.Level = logrus.InfoLevel
			break
		default:
			c.Logger.Level = logrus.DebugLevel
			break
		}
	}

	c.loggerFields = logrus.Fields{
		"command_name": c.Name,
	}

	c.waitGroup = &sync.WaitGroup{}
}

// GetStderr returns error output from the command
func (c *Command) GetStderr() string {
	return c.stderr
}

// GetStdout returns standard output from the command
func (c *Command) GetStdout() string {
	return c.stdout
}

// Run runs the given command
func (c *Command) Run() error {
	c.initialize()

	c.Logger.WithFields(c.loggerFields).Info("Running command")

	var err = c.run()
	if err == nil {
		c.Logger.WithFields(c.loggerFields).Info("Command succeeded")
	} else {
		c.Logger.WithFields(c.loggerFields).Warn("Command execution failed")
	}

	return err
}

func (c *Command) addValidationError(errorMessage string) {
	c.validationErrors = append(c.validationErrors, errorMessage)
}

func (c *Command) prepareRun() error {
	var err error

	if c.validate() {
		c.cmd = exec.Command(c.Executable, c.Arguments...)

		if c.WorkingDirectory != "" {
			var fs = filesystem.Filesystem{
				Logger:    c.Logger,
				Verbosity: c.Verbosity,
			}
			c.cmd.Dir, err = fs.BuildAbsolutePathFromHome(c.WorkingDirectory)
			if err != nil {
				return err
			}
			c.Logger.WithFields(c.loggerFields).Debug("Set working directory to " + c.cmd.Dir)
		}
		c.Logger.WithFields(c.loggerFields).Debug("Command: " + c.Executable + " " + strings.Join(c.Arguments, " "))

		// Set up stdout & stderr capture
		var stdoutPipe, stderrPipe io.ReadCloser

		stdoutPipe, err = c.cmd.StdoutPipe()
		if err != nil {
			c.Logger.WithFields(c.loggerFields).Warn("Could not create stdout pipe")
			return err
		}
		var stdoutScanner = bufio.NewScanner(stdoutPipe)
		c.waitGroup.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			var s string
			for stdoutScanner.Scan() {
				s = stdoutScanner.Text()
				c.stdout += s + "\n"
				if c.Verbosity >= 3 {
					c.Logger.WithFields(c.loggerFields).Info(s)
				}
			}
		}(c.waitGroup)

		stderrPipe, err = c.cmd.StderrPipe()
		if err != nil {
			c.Logger.WithFields(c.loggerFields).Warn("Could not create stderr pipe")
			return err
		}
		var stderrScanner = bufio.NewScanner(stderrPipe)
		c.waitGroup.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			var s string
			for stderrScanner.Scan() {
				s = stderrScanner.Text()
				c.stderr += s + "\n"
				if c.Verbosity >= 3 {
					c.Logger.WithFields(c.loggerFields).Warn(s)
				}
			}
		}(c.waitGroup)
	} else {
		c.Logger.WithFields(c.loggerFields).Warn("Command validation failed")
		err = errors.New("Command validation failed")
	}
	return err
}

func (c *Command) run() error {
	var err = c.prepareRun()

	if err == nil {
		err = c.cmd.Start()
		if err == nil {
			err = c.cmd.Wait()
			c.waitGroup.Wait()
		} else {
			c.Logger.Warn("Could not start process")
		}
	}

	return err
}

func (c *Command) validate() bool {
	if c.Name == "" {
		c.addValidationError("Name property is required")
	}
	if c.Executable == "" {
		c.addValidationError("Executable must be specified")
	}
	return len(c.validationErrors) == 0
}
