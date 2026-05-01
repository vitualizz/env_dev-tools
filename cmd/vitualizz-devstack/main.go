package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vitualizz/vitualizz-devstack/internal/config"
	"github.com/vitualizz/vitualizz-devstack/internal/domain/entities"
	"github.com/vitualizz/vitualizz-devstack/internal/infrastructure/installers"
	"github.com/vitualizz/vitualizz-devstack/internal/ui/components"
	"github.com/vitualizz/vitualizz-devstack/i18n/locales"
)

//go:embed config/tools.yaml
var embeddedToolsYAML []byte

//go:embed all:config/kitty/*
var embeddedKitty embed.FS

//go:embed all:config/zsh/*
var embeddedZsh embed.FS

func main() {
	// Check for flags
	ciMode := false
	selfDestruct := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--ci":
			ciMode = true
		case "--self-destruct":
			selfDestruct = true
		}
	}

	// If self-destruct mode, register cleanup to delete binary on exit
	binPath := ""
	if selfDestruct {
		var err error
		binPath, err = os.Executable()
		if err != nil {
			binPath = ""
		}
		defer func() {
			if binPath != "" {
				os.Remove(binPath)
			}
		}()
	}

	// Determine config source: env override or embedded
	var configPath string
	var configDir string

	if envPath := os.Getenv("DEVSTACK_CONFIG"); envPath != "" {
		// Development/testing: use external config file
		configPath = envPath
		configDir = filepath.Dir(envPath)
	} else {
		// Production: extract embedded config to temp directory
		var err error
		configDir, err = extractEmbeddedConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting config: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(configDir)
		configPath = filepath.Join(configDir, "tools.yaml")
	}

	repo, err := config.NewToolRepository(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	installer := installers.NewToolInstaller()
	installer.SetConfigDir(configDir)

	i18n := locales.NewI18nSimple()
	inDocker := entities.IsDocker()

	if ciMode {
		runCI(repo, installer, inDocker)
		return
	}

	runTUI(repo, installer, i18n)
}

// extractEmbeddedConfig extracts embedded config files to a temp directory
// and returns the path to the temp directory.
// Structure: tmpDir/tools.yaml, tmpDir/kitty/*, tmpDir/zsh/*
func extractEmbeddedConfig() (string, error) {
	tmpDir, err := os.MkdirTemp("", "devstack-config-*")
	if err != nil {
		return "", err
	}

	// Extract tools.yaml directly to tmpDir root
	if err := os.WriteFile(filepath.Join(tmpDir, "tools.yaml"), embeddedToolsYAML, 0o644); err != nil {
		return "", err
	}

	// Extract kitty configs to tmpDir/kitty/
	kittyDir := filepath.Join(tmpDir, "kitty")
	if err := os.MkdirAll(kittyDir, 0o755); err != nil {
		return "", err
	}

	kittyFiles, err := embeddedKitty.ReadDir("config/kitty")
	if err != nil {
		return "", err
	}
	for _, f := range kittyFiles {
		if f.IsDir() {
			continue
		}
		data, err := embeddedKitty.ReadFile(filepath.Join("config", "kitty", f.Name()))
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(kittyDir, f.Name()), data, 0o644); err != nil {
			return "", err
		}
	}

	// Extract zsh configs to tmpDir/zsh/
	zshDir := filepath.Join(tmpDir, "zsh")
	if err := os.MkdirAll(zshDir, 0o755); err != nil {
		return "", err
	}

	zshFiles, err := embeddedZsh.ReadDir("config/zsh")
	if err != nil {
		return "", err
	}
	for _, f := range zshFiles {
		if f.IsDir() {
			continue
		}
		data, err := embeddedZsh.ReadFile(filepath.Join("config", "zsh", f.Name()))
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(zshDir, f.Name()), data, 0o644); err != nil {
			return "", err
		}
	}

	return tmpDir, nil
}

// =============================================================================
// CI Mode — headless installation
// =============================================================================

func runCI(repo *config.ToolRepository, installer *installers.ToolInstaller, inDocker bool) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│  Vitualizz DevStack — CI Mode               │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	if inDocker {
		fmt.Println("  🐳 Docker environment detected")
		fmt.Println("  Skipping display tools (kitty, docker, etc.)")
		fmt.Println()
	}

	allTools := repo.GetAll()
	if inDocker {
		allTools = entities.FilterDockerIncompatible(allTools)
	}

	total := len(allTools)
	fmt.Printf("  Installing %d tools...\n\n", total)

	var success, failed int
	for _, tool := range allTools {
		if !tool.HasInstallCommand() {
			fmt.Printf("  ⊘ %s (bundle)\n", tool.Name)
			continue
		}

		// Check if already installed
		installed, _ := installer.IsInstalled(&tool)
		if installed {
			success++
			fmt.Printf("  ✓ %s (already installed)\n", tool.Name)
			continue
		}

		// Install
		fmt.Printf("  → %s... ", tool.Name)
		result, err := installer.Install(&tool)
		if err != nil || !result.Success {
			failed++
			fmt.Println("✗")
			if result != nil && result.Message != "" {
				fmt.Printf("    └─ %s\n", truncate(result.Message, 60))
			}
		} else {
			success++
			fmt.Println("✓")
		}
	}

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Printf("│  Total: %d  |  ✓ %d  |  ✗ %d           \n", total, success, failed)
	fmt.Println("└─────────────────────────────────────────────┘")

	if failed > 0 {
		fmt.Println()
		fmt.Println("Failed tools:")
		for _, tool := range allTools {
			if !tool.HasInstallCommand() {
				continue
			}
			installed, _ := installer.IsInstalled(&tool)
			if !installed {
				fmt.Printf("  ✗ %s\n", tool.Name)
			}
		}
		os.Exit(1)
	}
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// =============================================================================
// TUI Mode — interactive installation
// =============================================================================

func runTUI(repo *config.ToolRepository, installer *installers.ToolInstaller, i18n *locales.I18nSimple) {
	app := components.NewApp(repo, installer, i18n)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
