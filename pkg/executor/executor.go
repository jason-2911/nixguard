// Package executor provides safe command execution with whitelisting.
// Used by the privileged agent to ensure only approved commands run.
package executor

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Result holds command execution output.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Safe executes commands only if they are in the whitelist.
type Safe struct {
	log     *slog.Logger
	allowed map[string]bool
}

// NewSafe creates an executor with a command whitelist.
func NewSafe(log *slog.Logger, allowedCommands []string) *Safe {
	allowed := make(map[string]bool, len(allowedCommands))
	for _, cmd := range allowedCommands {
		allowed[cmd] = true
	}
	return &Safe{log: log, allowed: allowed}
}

// Run executes a command if the binary is whitelisted.
func (s *Safe) Run(ctx context.Context, name string, args ...string) (*Result, error) {
	if !s.allowed[name] {
		return nil, fmt.Errorf("command %q not in whitelist", name)
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, name, args...)

	s.log.Debug("executing command",
		slog.String("cmd", name),
		slog.String("args", strings.Join(args, " ")),
	)

	stdout, err := cmd.Output()
	duration := time.Since(start)

	result := &Result{
		Stdout:   string(stdout),
		Duration: duration,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		result.Stderr = string(exitErr.Stderr)
		s.log.Warn("command failed",
			slog.String("cmd", name),
			slog.Int("exit_code", result.ExitCode),
			slog.String("stderr", result.Stderr),
		)
		return result, fmt.Errorf("command %s exited with %d: %s", name, result.ExitCode, result.Stderr)
	}

	if err != nil {
		return nil, fmt.Errorf("command %s: %w", name, err)
	}

	s.log.Debug("command completed",
		slog.String("cmd", name),
		slog.Duration("duration", duration),
	)

	return result, nil
}

// RunRaw executes a command and returns raw stdout bytes.
func (s *Safe) RunRaw(ctx context.Context, name string, args ...string) ([]byte, error) {
	result, err := s.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	return []byte(result.Stdout), nil
}
