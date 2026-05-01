package main

import (
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

func main() {
	// Check for --ci flag
	ciMode := false
	for _, arg := range os.Args[1:] {
		if arg == "--ci" {
			ciMode = true
		}
	}

	configPath := getConfigPath()

	repo, err := config.NewToolRepository(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	installer := installers.NewToolInstaller()
	configDir := resolveConfigDir(configPath)
	installer.SetConfigDir(configDir)

	i18n := locales.NewI18nSimple()
	inDocker := entities.IsDocker()

	if ciMode {
		runCI(repo, installer, inDocker)
		return
	}

	runTUI(repo, installer, i18n)
}

func getConfigPath() string {
	if envPath := os.Getenv("ENVSETUP_CONFIG"); envPath != "" {
		return envPath
	}
	return "config/tools.yaml"
}

func resolveConfigDir(configPath string) string {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		return ""
	}
	return filepath.Dir(abs)
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
