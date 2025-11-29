package discovery

import (
	"bytes"
	"context"
	"os/exec"
)

// CommandExecutor defines the interface for executing external commands.
// This interface allows for testability by enabling mock implementations.
type CommandExecutor interface {
	// Execute runs a command and returns stdout, stderr, and any error.
	Execute(cmd string, args ...string) (stdout, stderr string, err error)

	// ExecuteContext runs a command with context support for timeout/cancellation.
	ExecuteContext(ctx context.Context, cmd string, args ...string) (stdout, stderr string, err error)
}

// DefaultExecutor is the default implementation of CommandExecutor using os/exec.
type DefaultExecutor struct{}

// NewDefaultExecutor creates a new DefaultExecutor instance.
func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{}
}

// Execute runs a command and returns stdout, stderr, and any error.
func (e *DefaultExecutor) Execute(cmd string, args ...string) (string, string, error) {
	return e.ExecuteContext(context.Background(), cmd, args...)
}

// ExecuteContext runs a command with context support for timeout/cancellation.
func (e *DefaultExecutor) ExecuteContext(ctx context.Context, cmd string, args ...string) (string, string, error) {
	command := exec.CommandContext(ctx, cmd, args...)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return stdout.String(), stderr.String(), err
}
