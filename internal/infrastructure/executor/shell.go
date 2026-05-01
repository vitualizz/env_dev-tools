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
	// ToolName is set before each command for logging purposes.
	ToolName string
	// LogFunc is called with (toolName, command, output, err, duration) after each execution.
	LogFunc func(toolName, command string, output string, err error, duration time.Duration)
}

func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

// toolPaths returns common directories where dev tools install binaries.
func toolPaths() string {
	return "$HOME/.cargo/bin:$HOME/.local/bin:$HOME/.mise/bin:$HOME/go/bin:$HOME/.opencode/bin:/usr/local/bin"
}

// wrapCmd prepends environment variable exports and tool paths if needed.
func (e *ShellExecutor) wrapCmd(cmd string) string {
	// Always prepend tool paths to PATH so tools like cargo, mise, etc. are found
	wrapped := fmt.Sprintf("export PATH=%s:$PATH && ", toolPaths())

	// Inject env vars if command references them
	if len(e.EnvVars) > 0 {
		hasRef := false
		for _, env := range e.EnvVars {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 && strings.Contains(cmd, "$"+parts[0]) {
				hasRef = true
				break
			}
		}
		if hasRef {
			var exports []string
			for _, env := range e.EnvVars {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					exports = append(exports, "export "+env)
				}
			}
			wrapped += strings.Join(exports, " && ") + " && "
		}
	}

	return wrapped + cmd
}

func (e *ShellExecutor) Execute(cmd string) (string, error) {
	start := time.Now()

	sh, c := shellArgs()
	command := exec.Command(sh, c, e.wrapCmd(cmd))
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	duration := time.Since(start)

	out := stdout.String()
	errOut := stderr.String()

	if e.LogFunc != nil {
		var combinedErr error
		if err != nil {
			combinedErr = fmt.Errorf("%s (stderr: %s)", err, errOut)
		}
		e.LogFunc(e.ToolName, cmd, out+"\n"+errOut, combinedErr, duration)
	}

	if err != nil {
		return "", fmt.Errorf("command failed: %w\n--- stderr ---\n%s", err, errOut)
	}

	return fmt.Sprintf("[%dms] %s", duration.Milliseconds(), out), nil
}

const installTimeout = 3 * time.Minute

func (e *ShellExecutor) ExecuteWithOutput(cmd string) (string, error) {
	start := time.Now()
	sh, c := shellArgs()

	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, sh, c, e.wrapCmd(cmd))
	output, err := command.CombinedOutput()
	duration := time.Since(start)

	if e.LogFunc != nil {
		e.LogFunc(e.ToolName, cmd, string(output), err, duration)
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return string(output), fmt.Errorf("command timed out after %v", installTimeout)
		}
		return string(output), err
	}
	return string(output), nil
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
