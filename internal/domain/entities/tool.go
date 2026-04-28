package entities

import (
	"os"
	"os/exec"
	"slices"
	"strings"
)

// =============================================================================
// Category
// =============================================================================

type Category string

const (
	CategoryTerminal   Category = "terminal"
	CategoryShell     Category = "shell"
	CategoryEditor   Category = "editor"
	CategoryTools    Category = "tools"
	CategoryContainer Category = "container"
	CategoryFonts   Category = "fonts"
	CategoryTheme  Category = "theme"
)

// AllCategories returns all tool categories in display order.
func AllCategories() []Category {
	return []Category{
		CategoryTerminal,
		CategoryShell,
		CategoryEditor,
		CategoryTools,
		CategoryContainer,
		CategoryFonts,
	}
}

func (c Category) String() string { return string(c) }

// IsValid checks if the category is a known tool category.
func (c Category) IsValid() bool {
	return slices.Contains(AllCategories(), c)
}

// =============================================================================
// Distro
// =============================================================================

// Distro represents a supported Linux distribution.
type Distro string

const (
	DistroArch     Distro = "arch"
	DistroDebian   Distro = "debian"
	DistroFedora   Distro = "fedora"
	DistroSuse     Distro = "suse"
	DistroAlpine   Distro = "alpine"
	DistroBrew     Distro = "brew"
	DistroAll      Distro = "all"
	DistroFallback Distro = "fallback"
)

// DistroDetectionOrder is the order in which distros are checked for fallbacks.
var DistroDetectionOrder = []Distro{DistroArch, DistroDebian, DistroFedora, DistroSuse, DistroAlpine, DistroBrew, DistroFallback}

// DetectDistro detects the current Linux distribution.
func DetectDistro() Distro {
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		lower := strings.ToLower(string(data))
		switch {
		case strings.Contains(lower, "arch"):
			return DistroArch
		case strings.Contains(lower, "debian") || strings.Contains(lower, "ubuntu"):
			return DistroDebian
		case strings.Contains(lower, "fedora") || strings.Contains(lower, "rhel") ||
			strings.Contains(lower, "rocky") || strings.Contains(lower, "almalinux"):
			return DistroFedora
		case strings.Contains(lower, "opensuse") || strings.Contains(lower, "sles"):
			return DistroSuse
		case strings.Contains(lower, "alpine"):
			return DistroAlpine
		}
	}

	// Package manager fallbacks
	cmds := []struct {
		distro Distro
		check string
	}{
		{DistroArch, "pacman"},
		{DistroDebian, "apt-get"},
		{DistroFedora, "dnf"},
		{DistroSuse, "zypper"},
		{DistroAlpine, "apk"},
		{DistroBrew, "brew"},
	}
	for _, c := range cmds {
		if _, err := exec.LookPath(c.check); err == nil {
			return c.distro
		}
	}

	return ""
}

// =============================================================================
// Theme
// =============================================================================

type Theme struct {
	Name        string      `json:"name" yaml:"name"`
	DisplayName string     `json:"display_name" yaml:"display_name"`
	Category   Category   `json:"category" yaml:"category"`
	Colors    ThemeColors `json:"colors" yaml:"colors"`
}

type ThemeColors struct {
	Background  string `json:"background" yaml:"background"`
	Foreground string `json:"foreground" yaml:"foreground"`
	Normal     ThemeColorSet `json:"normal" yaml:"normal"`
	Bright    ThemeColorSet `json:"bright" yaml:"bright"`
}

type ThemeColorSet struct {
	Black   string `json:"black" yaml:"black"`
	Red     string `json:"red" yaml:"red"`
	Green   string `json:"green" yaml:"green"`
	Yellow  string `json:"yellow" yaml:"yellow"`
	Blue    string `json:"blue" yaml:"blue"`
	Magenta string `json:"magenta" yaml:"magenta"`
	Cyan    string `json:"cyan" yaml:"cyan"`
	White   string `json:"white" yaml:"white"`
}

// =============================================================================
// Tool
// =============================================================================

// Tool represents a single tool or tool bundle.
type Tool struct {
	// Core identity
	Name        string   `json:"name" yaml:"name"`
	Category   Category `json:"category" yaml:"category"`
	Description string  `json:"description" yaml:"description"`

	// Installation commands per distro
	Install   map[Distro]string `json:"install" yaml:"install"`
	Uninstall map[Distro]string `json:"uninstall" yaml:"uninstall"`
	Check     string           `json:"check" yaml:"check"`

	// Dependencies (tools that must be installed first)
	DependsOn []string `json:"depends_on" yaml:"depends_on"`

	// Metadata
	SourceURL    string   `json:"source_url" yaml:"source_url"`
	Version     string   `json:"version" yaml:"version"`
	Alternatives []string `json:"alternatives" yaml:"alternatives"`

	// UI state
	Enabled bool `json:"enabled" yaml:"enabled"`
	Required bool `json:"required" yaml:"required"`
}

// GetInstallCmd returns the install command for the given distro.
// Tries: exact match → "all" → other distros in detection order.
func (t *Tool) GetInstallCmd(distro Distro) string {
	if cmd, ok := t.Install[distro]; ok && cmd != "" {
		return cmd
	}
	if cmd, ok := t.Install[DistroAll]; ok && cmd != "" {
		return cmd
	}
	for _, d := range DistroDetectionOrder {
		if d == distro {
			continue
		}
		if cmd, ok := t.Install[d]; ok && cmd != "" {
			return cmd
		}
	}
	return ""
}

// GetUninstallCmd returns the uninstall command for the given distro.
func (t *Tool) GetUninstallCmd(distro Distro) string {
	if cmd, ok := t.Uninstall[distro]; ok && cmd != "" {
		return cmd
	}
	if cmd, ok := t.Uninstall[DistroAll]; ok && cmd != "" {
		return cmd
	}
	for _, d := range DistroDetectionOrder {
		if d == distro {
			continue
		}
		if cmd, ok := t.Uninstall[d]; ok && cmd != "" {
			return cmd
		}
	}
	return ""
}

// HasInstallCommand returns true if the tool has any install command defined.
func (t *Tool) HasInstallCommand() bool {
	for _, cmd := range t.Install {
		if cmd != "" {
			return true
		}
	}
	return false
}

// HasUninstallCommand returns true if the tool has any uninstall command defined.
func (t *Tool) HasUninstallCommand() bool {
	for _, cmd := range t.Uninstall {
		if cmd != "" {
			return true
		}
	}
	return false
}

// IsBundle returns true if this tool is a bundle (group container).
// Bundles have empty install commands and no dependencies.
func (t *Tool) IsBundle() bool {
	return !t.HasInstallCommand() && len(t.DependsOn) == 0
}

// =============================================================================
// InstallResult
// =============================================================================

type InstallResult struct {
	ToolName   string `json:"tool_name"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	DurationMs int64  `json:"duration_ms"`
	Distro    Distro `json:"distro,omitempty"`
}

// =============================================================================
// Interfaces
// =============================================================================

type Installer interface {
	Install(tool *Tool) (*InstallResult, error)
	Uninstall(tool *Tool) (*InstallResult, error)
	IsInstalled(tool *Tool) (bool, error)
}

type Executor interface {
	Execute(cmd string) (string, error)
	ExecuteWithOutput(cmd string) (string, error)
}