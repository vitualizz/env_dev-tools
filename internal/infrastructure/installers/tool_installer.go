package installers

import (
	"strings"
	"time"

	"github.com/vitualizz/vitualizz-devstack/internal/domain/entities"
	"github.com/vitualizz/vitualizz-devstack/internal/infrastructure/executor"
)

// ToolInstaller handles installation and uninstallation of tools.
type ToolInstaller struct {
	exec     *executor.ShellExecutor
	distro   entities.Distro
	configDir string // path to config directory (for $DEVSTACK_CONFIG)
}

// NewToolInstaller creates a new installer and auto-detects the distro.
func NewToolInstaller() *ToolInstaller {
	return &ToolInstaller{
		exec:   executor.NewShellExecutor(),
		distro: entities.DetectDistro(),
	}
}

// NewToolInstallerWithDistro creates a new installer for a specific distro.
func NewToolInstallerWithDistro(distro entities.Distro) *ToolInstaller {
	return &ToolInstaller{
		exec:   executor.NewShellExecutor(),
		distro: distro,
	}
}

// SetConfigDir sets the config directory path for $DEVSTACK_CONFIG resolution.
func (i *ToolInstaller) SetConfigDir(dir string) {
	i.configDir = dir
	if dir != "" {
		i.exec.EnvVars = []string{"DEVSTACK_CONFIG=" + dir}
	}
}

// Distro returns the detected distro.
func (i *ToolInstaller) Distro() entities.Distro {
	return i.distro
}

// Install executes the install command for the given tool.
func (i *ToolInstaller) Install(tool *entities.Tool) (*entities.InstallResult, error) {
	start := time.Now()

	result := &entities.InstallResult{
		ToolName: tool.Name,
		Distro:   i.distro,
	}

	cmd := tool.GetInstallCmd(i.distro)
	if cmd == "" {
		result.Success = false
		result.Message = "no install command available for distro: " + string(i.distro)
		return result, nil
	}

	output, err := i.exec.ExecuteWithOutput(cmd)
	result.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Success = false
		result.Message = formatError(output, err)
		return result, err
	}

	result.Success = true
	result.Message = cleanOutput(output)
	return result, nil
}

// Uninstall executes the uninstall command for the given tool.
func (i *ToolInstaller) Uninstall(tool *entities.Tool) (*entities.InstallResult, error) {
	start := time.Now()

	result := &entities.InstallResult{
		ToolName: tool.Name,
		Distro:   i.distro,
	}

	cmd := tool.GetUninstallCmd(i.distro)
	if cmd == "" {
		result.Success = false
		result.Message = "no uninstall command available for distro: " + string(i.distro)
		return result, nil
	}

	output, err := i.exec.ExecuteWithOutput(cmd)
	result.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Success = false
		result.Message = formatError(output, err)
		return result, err
	}

	result.Success = true
	result.Message = cleanOutput(output)
	return result, nil
}

// IsInstalled checks if a tool is already installed.
func (i *ToolInstaller) IsInstalled(tool *entities.Tool) (bool, error) {
	if tool.Check == "" {
		return false, nil
	}

	output, err := i.exec.ExecuteWithOutput(tool.Check)
	if err != nil {
		return false, nil
	}

	return strings.TrimSpace(output) != "", nil
}

// formatError returns a user-friendly error message.
func formatError(output string, err error) string {
	if output != "" {
		return strings.TrimSpace(output)
	}
	return err.Error()
}

// cleanOutput removes common noise from command output.
func cleanOutput(output string) string {
	lines := strings.Split(output, "\n")
	var clean []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, "\r") {
			line = strings.SplitN(line, "\r", 2)[1]
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		clean = append(clean, line)
	}

	if len(clean) == 0 {
		return ""
	}

	return strings.Join(clean, "\n")
}

// Compile-time interface check
var _ entities.Installer = (*ToolInstaller)(nil)
