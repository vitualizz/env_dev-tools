package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vitualizz/envsetup/internal/domain/entities"
	"github.com/vitualizz/envsetup/internal/domain/interfaces"
	"github.com/vitualizz/envsetup/internal/ui/models"
	"github.com/vitualizz/envsetup/internal/usecases"
	"github.com/vitualizz/envsetup/i18n/locales"
)

type App struct {
	model        *models.AppModel
	repo         interfaces.ToolRepository
	listTools    *usecases.ListToolsUseCase
	installTool  *usecases.InstallToolUseCase
	batchInstall *usecases.BatchInstallUseCase
	checkStatus  *usecases.BatchCheckStatusUseCase
	i18n         *locales.I18nSimple
}

func NewApp(repo interfaces.ToolRepository, installer interfaces.InstallerPort, i18n *locales.I18nSimple) *App {
	return &App{
		model:        models.NewAppModel(),
		repo:         repo,
		listTools:    usecases.NewListToolsUseCase(repo),
		installTool:  usecases.NewInstallToolUseCase(installer, repo),
		batchInstall: usecases.NewBatchInstallUseCase(installer, repo),
		checkStatus:  usecases.NewBatchCheckStatusUseCase(installer),
		i18n:         i18n,
	}
}

func (a *App) Init() tea.Cmd {
	tools := a.listTools.GetMainTools()
	a.model.SetTools(tools)
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKey(msg.String())
	case tea.WindowSizeMsg:
		return a, nil
	case statusCheckCompleteMsg:
		a.model.ToolStatus = msg.status
		a.model.StatusChecked = true
		return a, nil
	case installCompleteMsg:
		a.model.ViewState = models.StateMainMenu
		results := make([]entities.InstallResult, len(msg.results))
		for i, r := range msg.results {
			results[i] = *r
		}
		a.model.Results = results
		a.model.IsLoading = false
		return a, nil
	case errMsg:
		a.model.IsLoading = false
		return a, nil
	}
	return a, nil
}

func (a *App) handleKey(key string) (tea.Model, tea.Cmd) {
	if a.model.IsLoading {
		return a, nil
	}

	switch a.model.ViewState {
	case models.StateLanguageSelect:
		return a.handleLanguageSelect(key)
	case models.StateThemeSelect:
		return a.handleThemeSelect(key)
	case models.StateMainMenu:
		return a.handleMainMenu(key)
	case models.StateToolList:
		return a.handleToolList(key)
	case models.StateSettings:
		return a.handleSettings(key)
	case models.StateAbout:
		if key == "esc" {
			a.model.ViewState = models.StateMainMenu
		}
	}
	return a, nil
}

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
		return a, nil
	}
	return a, nil
}

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
		tools := a.listTools.GetMainTools()
		a.model.SetTools(tools)
	case "esc":
		a.model.SettingsChoice = 0
		a.model.ViewState = models.StateLanguageSelect
	}
	return a, nil
}

func (a *App) handleMainMenu(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if a.model.MainMenuChoice > 0 {
			a.model.MainMenuChoice--
		}
	case "down":
		if a.model.MainMenuChoice < 3 {
			a.model.MainMenuChoice++
		}
	case "enter":
		switch a.model.MainMenuChoice {
		case 0:
			a.model.ViewState = models.StateToolList
			a.model.StatusChecked = false
			return a, a.runStatusCheck()
		case 1:
			a.model.ViewState = models.StateSettings
		case 2:
			a.model.ViewState = models.StateAbout
		case 3:
			return a, tea.Quit
		}
	case "q", "ctrl+c":
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) handleToolList(key string) (tea.Model, tea.Cmd) {
	filteredTools := a.model.GetFilteredTools()
	maxChoice := len(filteredTools) - 1

	switch key {
	case "up":
		if a.model.ToolChoice > 0 {
			a.model.ToolChoice--
		}
	case "down":
		if a.model.ToolChoice < maxChoice {
			a.model.ToolChoice++
		}
	case " ":
		if len(filteredTools) > 0 {
			tool := filteredTools[a.model.ToolChoice]
			a.model.ToggleToolSelection(tool.Name)
		}
	case "enter":
		selected := a.model.GetSelectedToolsList()
		if len(selected) > 0 {
			a.model.IsLoading = true
			a.model.LoadingMessage = a.i18n.T("installing")
			return a, a.executeInstall(selected)
		}
	case "esc":
		a.model.ViewState = models.StateMainMenu
		a.model.FilterQuery = ""
	}
	return a, nil
}

func (a *App) handleSettings(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if a.model.SettingsChoice > 0 {
			a.model.SettingsChoice--
		}
	case "down":
		if a.model.SettingsChoice < 2 {
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

func (a *App) runStatusCheck() tea.Cmd {
	return func() tea.Msg {
		status := a.checkStatus.Execute(a.model.Tools)
		return statusCheckCompleteMsg{status: status}
	}
}

func (a *App) executeInstall(tools []entities.Tool) tea.Cmd {
	return func() tea.Msg {
		toolsPtr := make([]*entities.Tool, len(tools))
		for i := range tools {
			toolsPtr[i] = &tools[i]
		}
		results := a.batchInstall.Execute(toolsPtr)
		return installCompleteMsg{results}
	}
}

func (a *App) View() string {
	switch a.model.ViewState {
	case models.StateLanguageSelect:
		return a.languageSelectView()
	case models.StateThemeSelect:
		return a.themeSelectView()
	case models.StateMainMenu:
		return a.mainMenuView()
	case models.StateToolList:
		return a.toolListView()
	case models.StateInstalling:
		return a.loadingView()
	case models.StateSettings:
		return a.settingsView()
	case models.StateAbout:
		return a.aboutView()
	}
	return ""
}

func (a *App) languageSelectView() string {
	languages := []struct {
		code string
		name string
	}{
		{"es", "Español"},
		{"en", "English"},
	}

	output := headerView(a.i18n.T("language")) + "\n\n"

	for i, lang := range languages {
		if i == a.model.SettingsChoice {
			output += selectedView("► " + lang.name) + "\n"
		} else {
			output += unselectedView("  " + lang.name) + "\n"
		}
	}

	output += "\n" + footerView("↑↓: " + a.i18n.T("select_option") + " | Enter: " + a.i18n.T("confirm"))
	return output
}

func (a *App) themeSelectView() string {
	themes := a.repo.GetThemes()

	output := headerView(a.i18n.T("select_category")) + "\n\n"

	if len(themes) == 0 {
		output += warningView(a.i18n.T("no_tools_found"))
		return output
	}

	for i, theme := range themes {
		if i == a.model.ToolChoice {
			output += selectedView("► " + theme.DisplayName) + "\n"
		} else {
			output += unselectedView("  " + theme.DisplayName) + "\n"
		}
	}

	output += "\n" + footerView("↑↓: " + a.i18n.T("select_option") + " | Enter: " + a.i18n.T("confirm") + " | Esc: " + a.i18n.T("back"))
	return output
}

func (a *App) mainMenuView() string {
	items := []string{
		"📦 " + a.i18n.T("install"),
		"⚙️ " + a.i18n.T("settings"),
		"ℹ️ " + a.i18n.T("about"),
		"🚪 " + a.i18n.T("exit"),
	}

	themeName := a.model.CurrentTheme
	if themeName == "" {
		themeName = "Nord"
	}

	output := headerView(a.i18n.T("welcome") + " - "+ themeName) + "\n\n"

	for i, item := range items {
		if i == a.model.MainMenuChoice {
			output += selectedView("► " + item) + "\n"
		} else {
			output += unselectedView("  " + item) + "\n"
		}
	}

	output += "\n" + footerView("↑↓: " + a.i18n.T("select_option") + " | Enter: " + a.i18n.T("confirm"))
	return output
}

func (a *App) toolListView() string {
	output := headerView(a.i18n.T("select_tools")) + "\n\n"

	filteredTools := a.model.GetFilteredTools()

	if len(filteredTools) == 0 {
		return output + warningView(a.i18n.T("no_tools_found"))
	}

	for i, tool := range filteredTools {
		selected := a.model.IsToolSelected(tool.Name)
		checkbox := "[ ]"
		if selected {
			checkbox = "[✓]"
		}

		var status string
		if !a.model.StatusChecked {
			status = toolDescView("⟳ " + a.i18n.T("checking"))
		} else if a.model.ToolStatus[tool.Name] {
			status = successView("● " + a.i18n.T("installed"))
		} else {
			status = notInstalledView("○ " + a.i18n.T("not_installed"))
		}

		if i == a.model.ToolChoice {
			if selected {
				output += selectedView("► "+checkbox+" ") + toolNameView(tool.Name) + " " + status + "\n"
			} else {
				output += selectedView("► "+checkbox+" ") + toolNameView(tool.Name) + " " + status + "\n"
			}
			if tool.Description != "" {
				output += toolDescView("   " + tool.Description) + "\n"
			}
		} else {
			if selected {
				output += successView("  "+checkbox+" ") + toolNameView(tool.Name) + " " + status + "\n"
			} else {
				output += unselectedView("  "+checkbox+" ") + toolNameView(tool.Name) + " " + status + "\n"
			}
		}
	}

	selectedCount := len(a.model.GetSelectedToolsList())
	footer := fmt.Sprintf("%s: %d | ↑↓: %s | Space: %s | Enter: %s | Esc: %s",
		a.i18n.T("total"), selectedCount, a.i18n.T("select_option"), a.i18n.T("select_tools"), a.i18n.T("install_all"), a.i18n.T("back"))

	output += "\n" + footerView(footer)
	return output
}

func (a *App) settingsView() string {
	themeName := a.model.CurrentTheme
	if themeName == "" {
		themeName = "Nord"
	}

	items := []string{
		"🌐 " + a.i18n.T("language") + ": " + a.model.CurrentLang,
		"🎨 " + a.i18n.T("select_category") + ": " + themeName,
		"⬅️ " + a.i18n.T("back"),
	}

	output := headerView(a.i18n.T("settings")) + "\n\n"

	for i, item := range items {
		if i == a.model.SettingsChoice {
			output += selectedView("► " + item) + "\n"
		} else {
			output += unselectedView("  " + item) + "\n"
		}
	}

	output += "\n" + footerView("↑↓: " + a.i18n.T("select_option") + " | Enter: " + a.i18n.T("confirm"))
	return output
}

func (a *App) aboutView() string {
	themeName := a.model.CurrentTheme
	if themeName == "" {
		themeName = "Nord"
	}

	output := headerView(a.i18n.T("about")) + "\n\n"
	output += toolNameView("EnvSetup") + " " + toolDescView("v1.0.0") + "\n\n"
	output += toolDescView("Theme: " + themeName) + "\n"
	output += toolDescView("Language: " + a.model.CurrentLang) + "\n\n"
	output += footerView("Press Esc to "+a.i18n.T("back"))
	return output
}

func (a *App) loadingView() string {
	return headerView(a.model.LoadingMessage)
}

func headerView(text string) string {
	return "\033[36m\033[1m" + text + "\033[0m"
}

func footerView(text string) string {
	return "\033[90m" + text + "\033[0m"
}

func selectedView(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func unselectedView(text string) string {
	return "\033[90m" + text + "\033[0m"
}

func toolNameView(text string) string {
	return "\033[38;5;212m" + text + "\033[0m"
}

func toolDescView(text string) string {
	return "\033[38;5;245m" + text + "\033[0m"
}

func successView(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func errorView(text string) string {
	return "\033[31m" + text + "\033[0m"
}

func warningView(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func notInstalledView(text string) string {
	return "\033[90m" + text + "\033[0m"
}

type statusCheckCompleteMsg struct {
	status map[string]bool
}

type installCompleteMsg struct {
	results []*entities.InstallResult
}

type errMsg struct {
	err error
}

var _ tea.Model = (*App)(nil)