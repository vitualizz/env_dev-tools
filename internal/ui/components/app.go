package components

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/vitualizz/vitualizz-devstack/internal/domain/entities"
	"github.com/vitualizz/vitualizz-devstack/internal/domain/interfaces"
	"github.com/vitualizz/vitualizz-devstack/internal/ui/models"
	"github.com/vitualizz/vitualizz-devstack/internal/ui/views"
	"github.com/vitualizz/vitualizz-devstack/internal/usecases"
	"github.com/vitualizz/vitualizz-devstack/i18n/locales"
)

type App struct {
	model        *models.AppModel
	repo         interfaces.ToolRepository
	installer    interfaces.InstallerPort
	batchInstall *usecases.BatchInstallUseCase
	i18n         *locales.I18nSimple
	renderer     *views.Renderer
	inDocker     bool // true when running inside a Docker container
}

func NewApp(repo interfaces.ToolRepository, installer interfaces.InstallerPort, i18n *locales.I18nSimple) *App {
	app := &App{
		model:        models.NewAppModel(),
		repo:         repo,
		installer:    installer,
		batchInstall: usecases.NewBatchInstallUseCase(installer, repo),
		i18n:         i18n,
		inDocker:     entities.IsDocker(),
	}
	app.model.InDocker = app.inDocker
	app.renderer = &views.Renderer{
		Model: app.model,
		Repo:  repo,
		I18n:  i18n,
	}
	return app
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKey(msg.String())
	case tea.WindowSizeMsg:
		return a, nil
	case progressUpdateMsg:
		a.model.UpdateProgress(msg.tool, msg.success, msg.message)
		if a.model.IsProgressDone() {
			a.model.IsLoading = false
			a.model.ViewState = models.StateThanks
		} else {
			if a.model.IsUninstallMode {
				return a, a.uninstallNext()
			}
			return a, a.installNext()
		}
		return a, nil
	}
	return a, nil
}

func (a *App) handleKey(key string) (tea.Model, tea.Cmd) {
	if key == "q" || key == "ctrl+c" {
		return a, tea.Quit
	}

	if a.model.IsLoading && a.model.ViewState != models.StateProgress {
		return a, nil
	}

	switch a.model.ViewState {
	case models.StateLanguageSelect:
		return a.handleLanguageSelect(key)
	case models.StateThemeSelect:
		return a.handleThemeSelect(key)
	case models.StateMainMenu:
		return a.handleMainMenu(key)
	case models.StateThanks:
		return a.handleThanks(key)
	case models.StateProgress:
		if key == "esc" {
			a.model.ViewState = models.StateMainMenu
			return a, nil
		}
		return a, nil
	case models.StateSettings:
		return a.handleSettings(key)
	case models.StateAbout:
		if key == "esc" {
			a.model.ViewState = models.StateMainMenu
		}
	}
	return a, nil
}

// =============================================================================
// Thanks (post-install/uninstall)
// =============================================================================

func (a *App) handleThanks(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "l":
		a.model.ShowLog = !a.model.ShowLog
	case "enter", " ", "esc", "q", "ctrl+c":
		return a, tea.Quit
	}
	return a, nil
}

// =============================================================================
// Language Select
// =============================================================================

func (a *App) handleLanguageSelect(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if a.model.SettingsChoice > 0 {
			a.model.SettingsChoice--
		}
	case "down":
		if a.model.SettingsChoice < 1 {
			a.model.SettingsChoice++
		}
	case "enter":
		if a.model.SettingsChoice == 0 {
			a.model.CurrentLang = "es"
			a.i18n.SetLanguage("es")
		} else {
			a.model.CurrentLang = "en"
			a.i18n.SetLanguage("en")
		}
		a.model.SettingsChoice = 0
		a.model.ViewState = models.StateThemeSelect
	}
	return a, nil
}

// =============================================================================
// Theme Select
// =============================================================================

func (a *App) handleThemeSelect(key string) (tea.Model, tea.Cmd) {
	themes := a.repo.GetThemes()
	maxChoice := len(themes) - 1

	switch key {
	case "up":
		if a.model.ToolChoice > 0 {
			a.model.ToolChoice--
		}
	case "down":
		if a.model.ToolChoice < maxChoice {
			a.model.ToolChoice++
		}
	case "enter":
		if len(themes) > 0 {
			a.model.CurrentTheme = themes[a.model.ToolChoice].Name
		}
		a.model.ToolChoice = 0
		a.model.ViewState = models.StateMainMenu
	case "esc":
		a.model.SettingsChoice = 0
		a.model.ViewState = models.StateLanguageSelect
	}
	return a, nil
}

// =============================================================================
// Main Menu
// =============================================================================

func (a *App) handleMainMenu(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if a.model.MainMenuChoice > 0 {
			a.model.MainMenuChoice--
		}
	case "down":
		if a.model.MainMenuChoice < a.renderer.MenuItemCount()-1 {
			a.model.MainMenuChoice++
		}
	case "enter":
		switch a.model.MainMenuChoice {
		case 0: // Install All
			allTools := a.repo.GetAll()
			// Filter out Docker-incompatible tools when running in container
			if a.inDocker {
				allTools = entities.FilterDockerIncompatible(allTools)
			}
			ordered := a.batchInstall.ResolveInstallOrder(allTools)
			a.model.StartProgress(ordered)
			a.model.IsLoading = true
			return a, a.installNext()
		case 1: // Uninstall All
			installed := a.getInstalledTools()
			if len(installed) == 0 {
				// Nothing to uninstall, show a quick thanks
				a.model.StartUninstallProgress(nil)
				a.model.ProgressIdx = 0 // already done
				a.model.ViewState = models.StateThanks
				a.model.IsLoading = false
				return a, nil
			}
			a.model.StartUninstallProgress(installed)
			a.model.IsLoading = true
			return a, a.uninstallNext()
		case 2: // Settings
			a.model.ViewState = models.StateSettings
		case 3: // About
			a.model.ViewState = models.StateAbout
		case 4: // Exit
			return a, tea.Quit
		}
	}
	return a, nil
}

// =============================================================================
// Progress (install)
// =============================================================================

func (a *App) installNext() tea.Cmd {
	tool := a.model.GetCurrentProgressTool()
	toolPtr := &tool
	return func() tea.Msg {
		installed, _ := a.installer.IsInstalled(toolPtr)
		if installed {
			return progressUpdateMsg{tool: tool, success: true, message: "already installed"}
		}
		result, err := a.installer.Install(toolPtr)
		if err != nil {
			return progressUpdateMsg{tool: tool, success: false, message: err.Error()}
		}
		return progressUpdateMsg{tool: tool, success: result.Success, message: result.Message}
	}
}

// =============================================================================
// Progress (uninstall)
// =============================================================================

func (a *App) uninstallNext() tea.Cmd {
	tool := a.model.GetCurrentProgressTool()
	toolPtr := &tool
	return func() tea.Msg {
		installed, _ := a.installer.IsInstalled(toolPtr)
		if !installed {
			return progressUpdateMsg{tool: tool, success: true, message: "not installed"}
		}
		result, err := a.installer.Uninstall(toolPtr)
		if err != nil {
			return progressUpdateMsg{tool: tool, success: false, message: err.Error()}
		}
		return progressUpdateMsg{tool: tool, success: result.Success, message: result.Message}
	}
}

// =============================================================================
// Settings
// =============================================================================

func (a *App) handleSettings(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if a.model.SettingsChoice > 0 {
			a.model.SettingsChoice--
		}
	case "down":
		if a.model.SettingsChoice < a.renderer.SettingsItemCount()-1 {
			a.model.SettingsChoice++
		}
	case "enter":
		switch a.model.SettingsChoice {
		case 0:
			if a.model.CurrentLang == "es" {
				a.model.CurrentLang = "en"
				a.i18n.SetLanguage("en")
			} else {
				a.model.CurrentLang = "es"
				a.i18n.SetLanguage("es")
			}
		case 1:
			a.model.ViewState = models.StateThemeSelect
			a.model.ToolChoice = 0
		case 2:
			a.model.ViewState = models.StateMainMenu
		}
	case "esc":
		a.model.ViewState = models.StateMainMenu
	}
	return a, nil
}

// =============================================================================
// Helpers
// =============================================================================

func (a *App) getInstalledTools() []entities.Tool {
	allTools := a.repo.GetAll()
	if a.inDocker {
		allTools = entities.FilterDockerIncompatible(allTools)
	}
	var installed []entities.Tool
	for _, t := range allTools {
		if !t.HasInstallCommand() {
			continue
		}
		isInstalled, _ := a.installer.IsInstalled(&t)
		if isInstalled {
			installed = append(installed, t)
		}
	}
	return installed
}

// =============================================================================
// View — delegates to Renderer
// =============================================================================

func (a *App) View() string {
	return a.renderer.View()
}

// =============================================================================
// Messages
// =============================================================================

type progressUpdateMsg struct {
	tool    entities.Tool
	success bool
	message string
}

var _ tea.Model = (*App)(nil)
