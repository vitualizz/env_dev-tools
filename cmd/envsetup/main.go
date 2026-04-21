package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vitualizz/envsetup/internal/config"
	"github.com/vitualizz/envsetup/internal/infrastructure/installers"
	"github.com/vitualizz/envsetup/internal/ui/components"
	"github.com/vitualizz/envsetup/i18n/locales"
)

func main() {
	configPath := getConfigPath()

	repo, err := config.NewToolRepository(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	installer := installers.NewToolInstaller()
	i18n := locales.NewI18nSimple()

	app := components.NewApp(repo, installer, i18n)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

func getConfigPath() string {
	if envPath := os.Getenv("ENVSETUP_CONFIG"); envPath != "" {
		return envPath
	}
	return "config/tools.yaml"
}