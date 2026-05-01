package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ShellExecutor runs shell commands via /bin/sh or /bin/zsh.
type ShellExecutor struct {
	// EnvVars are prepended as exports to every command.
	EnvVars []string
}

func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

func (e *ShellExecutor) Execute(cmd string) (string, error) {
	start := time.Now()

	sh, c := shellArgs()
	command := exec.Command(sh, c, e.wrapCmd(cmd))
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return "", fmt.Errorf("command failed: %w - %s", err, stderr.String())
	}

	return fmt.Sprintf("[%dms] %s", duration, stdout.String()), nil
}

const installTimeout = 3 * time.Minute

func (e *ShellExecutor) ExecuteWithOutput(cmd string) (string, error) {
	sh, c := shellArgs()

	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, sh, c, e.wrapCmd(cmd))
	output, err := command.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return string(output), fmt.Errorf("command timed out after 10 minutes")
		}
		return string(output), err
	}
	return string(output), nil
}

// wrapCmd prepends environment variable exports if needed.
func (e *ShellExecutor) wrapCmd(cmd string) string {
	if len(e.EnvVars) == 0 {
		return cmd
	}

	// Only inject env vars if the command references them
	hasRef := false
	for _, env := range e.EnvVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && strings.Contains(cmd, "$"+parts[0]) {
			hasRef = true
			break
		}
	}
	if !hasRef {
		return cmd
	}

	var exports []string
	for _, env := range e.EnvVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			exports = append(exports, "export "+env)
		}
	}
	return strings.Join(exports, " && ") + " && " + cmd
}

// shellArgs returns the preferred shell and argument flag.
func shellArgs() (string, string) {
	if isZshAvailable() {
		return "/bin/zsh", "-c"
	}
	return "/bin/sh", "-c"
}

func isZshAvailable() bool {
	_, err := exec.LookPath("zsh")
	return err == nil
}
