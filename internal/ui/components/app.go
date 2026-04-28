package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vitualizz/envsetup/internal/domain/entities"
	"github.com/vitualizz/envsetup/internal/domain/interfaces"
	"github.com/vitualizz/envsetup/internal/ui/models"
	"github.com/vitualizz/envsetup/internal/usecases"
	"github.com/vitualizz/envsetup/i18n/locales"
)

type App struct {
	model        *models.AppModel
	repo        interfaces.ToolRepository
	installer   interfaces.InstallerPort
	batchInstall *usecases.BatchInstallUseCase
	checkStatus  *usecases.BatchCheckStatusUseCase
	i18n        *locales.I18nSimple
}

func NewApp(repo interfaces.ToolRepository, installer interfaces.InstallerPort, i18n *locales.I18nSimple) *App {
	return &App{
		model:        models.NewAppModel(),
		repo:        repo,
		installer:   installer,
		batchInstall: usecases.NewBatchInstallUseCase(installer, repo),
		checkStatus:  usecases.NewBatchCheckStatusUseCase(installer),
		i18n:        i18n,
	}
}

func (a *App) Init() tea.Cmd {
	pkgs := a.repo.GetPackages()
	a.model.SetPackages(pkgs)
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
	case progressUpdateMsg:
		a.model.UpdateProgress(msg.tool, msg.success, msg.message)
		if a.model.IsProgressDone() {
			a.model.IsLoading = false
			a.model.ViewState = models.StateThanks
		} else {
			return a, a.installNext()
		}
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
	// En cualquier estado, q/ctrl+c sale
	if key == "q" || key == "ctrl+c" {
		return a, tea.Quit
	}

	// En progreso, solo Esc para salir
	if a.model.ViewState == models.StateProgress && key == "esc" {
		a.model.ViewState = models.StateMainMenu
		return a, nil
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
	case models.StatePackageSelect:
		return a.handlePackageSelect(key)
	case models.StateThanks:
		return a.handleThanks(key)
	case models.StateProgress:
		// En progreso, solo mostrar
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
// Thanks (post-install)
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
// Package Selection
// =============================================================================
// Language
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
// Theme
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
		pkgs := a.repo.GetPackages()
		a.model.SetPackages(pkgs)
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
		if a.model.MainMenuChoice < 3 {
			a.model.MainMenuChoice++
		}
	case "enter":
		switch a.model.MainMenuChoice {
		case 0:
			a.model.ViewState = models.StatePackageSelect
			a.model.SelectedIdx = 0
			a.model.StatusChecked = false
			return a, a.runStatusCheck()
		case 1:
			a.model.ViewState = models.StateSettings
		case 2:
			a.model.ViewState = models.StateAbout
		case 3:
			return a, tea.Quit
		}
	}
	return a, nil
}

// =============================================================================
// Package Selection
// =============================================================================

func (a *App) handlePackageSelect(key string) (tea.Model, tea.Cmd) {
	// Solo paquetes visibles (sin Fuentes)
	var visible []interfaces.Package
	for _, pkg := range a.model.Packages {
		if pkg.Name != "fonts" {
			visible = append(visible, pkg)
		}
	}
	maxChoice := len(visible) - 1

	switch key {
	case "up":
		if a.model.SelectedIdx > 0 {
			a.model.SelectedIdx--
		}
	case "down":
		if a.model.SelectedIdx < maxChoice {
			a.model.SelectedIdx++
		}
	case "a": // seleccionar todo (incluye Fonts implícito)
		a.model.SelectAll()
	case "A": // deseleccionar todo
		a.model.DeselectAll()
	case " ":
		if len(visible) > 0 {
			// Mapear índice visible a índice real en Packages
			realIdx := a.visibleToRealIdx(a.model.SelectedIdx)
			a.model.TogglePackageSelection(realIdx)
		}
	case "enter", "c": // Continuar
		selectedPkgs := a.model.GetSelectedPackages()
		if len(selectedPkgs) > 0 {
			tools := a.model.GetSelectedTools()
			a.model.StartProgress(tools)
			a.model.IsLoading = true
			return a, a.installNext()
		}
	case "esc":
		a.model.ViewState = models.StateMainMenu
	}
	return a, nil
}

// visibleToRealIdx convierte índice visible a índice real en Packages.
func (a *App) visibleToRealIdx(visibleIdx int) int {
	vis := 0
	for i, pkg := range a.model.Packages {
		if pkg.Name != "fonts" {
			if vis == visibleIdx {
				return i
			}
			vis++
		}
	}
	return visibleIdx
}

// =============================================================================
// Progress (instalación paso a paso)
// =============================================================================

func (a *App) installNext() tea.Cmd {
	tool := a.model.GetCurrentProgressTool()
	toolPtr := &tool
	return func() tea.Msg {
		// Verificar si ya está instalado
		installed, _ := a.installer.IsInstalled(toolPtr)
		if installed {
			return progressUpdateMsg{tool: tool, success: true, message: "already installed"}
		}
		// Instalar
		result, err := a.installer.Install(toolPtr)
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

// =============================================================================
// Background Tasks
// =============================================================================

func (a *App) runStatusCheck() tea.Cmd {
	return func() tea.Msg {
		allTools := a.repo.GetAll()
		status := a.checkStatus.Execute(allTools)
		return statusCheckCompleteMsg{status: status}
	}
}

// =============================================================================
// View
// =============================================================================

func (a *App) View() string {
	switch a.model.ViewState {
	case models.StateLanguageSelect:
		return a.languageSelectView()
	case models.StateThemeSelect:
		return a.themeSelectView()
	case models.StateMainMenu:
		return a.mainMenuView()
	case models.StatePackageSelect:
		return a.packageSelectView()
	case models.StateThanks:
		return a.thanksView()
	case models.StateProgress:
		return a.progressView()
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

	output := headerView("Vitualizz Space - Idioma") + "\n\n"
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
	output := headerView("Vitualizz Space - Themes") + "\n\n"

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

	output := headerView("Vitualizz Space") + "\n"
	output += toolDescView("github.com/vitualizz · vitualizz.vercel.app") + "\n\n"
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

func (a *App) packageSelectView() string {
	// Filtrar: no mostrar Fuentes en la lista (se instala por defecto)
	var visiblePkgs []interfaces.Package
	for _, pkg := range a.model.Packages {
		if pkg.Name != "fonts" {
			visiblePkgs = append(visiblePkgs, pkg)
		}
	}

	output := headerView("Vitualizz Space") + "\n"
	output += toolDescView("github.com/vitualizz · vitualizz.vercel.app") + "\n\n"
	output += warningView("🔤 "+a.i18n.T("fonts")+": "+a.i18n.T("enabled")+"\n")

	if len(visiblePkgs) == 0 {
		return output + warningView(a.i18n.T("no_tools_found"))
	}

	// Adjust cursor si se fue al final
	if a.model.SelectedIdx >= len(visiblePkgs) {
		a.model.SelectedIdx = len(visiblePkgs) - 1
	}

	for i, pkg := range visiblePkgs {
		selected := pkg.Selected
		checkbox := "[ ]"
		if selected {
			checkbox = "[✓]"
		}

		suffix := fmt.Sprintf(" (%d)", len(pkg.Tools))
		label := pkg.Icon + " " + pkg.Label + suffix

		if i == a.model.SelectedIdx {
			output += selectedView("► "+checkbox+" ") + toolNameView(label) + "\n"
			// Descripción del paquete
			output += toolDescView("   "+pkg.Description) + "\n"
			// Lista de descripciones de tools
			var descs []string
			for _, t := range pkg.Tools {
				descs = append(descs, a.getToolDescription(t.Name))
			}
			output += toolDescView("   📦 "+strings.Join(descs, " · ")) + "\n"
		} else {
			if selected {
				output += successView("  "+checkbox+" ") + label + "\n"
			} else {
				output += unselectedView("  "+checkbox+" ") + label + "\n"
			}
		}
	}

	// Contar packages seleccionados (incluyendo Fonts implícito)
	selectedCount := len(a.model.GetSelectedPackages())
	footer := fmt.Sprintf("%s: %d | ↑↓: %s | Space: %s | A: all | Enter: %s | Esc: %s",
		a.i18n.T("total"), selectedCount, a.i18n.T("select_option"), a.i18n.T("select_tools"), a.i18n.T("continue"), a.i18n.T("back"))

	output += "\n" + footerView(footer)
	return output
}

func (a *App) thanksView() string {
	total, success, failed := a.model.GetProgressStats()

	output := headerView("✓ "+a.i18n.T("completed")) + "\n\n"
	output += toolDescView("github.com/vitualizz") + "\n"
	output += toolDescView("vitualizz.vercel.app") + "\n\n"

	output += fmt.Sprintf("%s: %d | ", toolNameView("Total"), total)
	output += fmt.Sprintf("%s: %d", successView("✓"), success)
	if failed > 0 {
		output += fmt.Sprintf(" | %s: %d\n\n", errorView("✗"), failed)
	} else {
		output += "\n\n"
	}

	// Mostrar éxitos
	for _, r := range a.model.ProgressResults {
		if r.Success {
			desc := a.getToolDescription(r.ToolName)
			output += successView("✓ ") + toolDescView(desc) + "\n"
		}
	}

	// Mostrar errores
	if failed > 0 {
		output += "\n" + errorView("✗ "+a.i18n.T("some_errors")) + "\n"
		for _, r := range a.model.ProgressResults {
			if !r.Success {
				desc := a.getToolDescription(r.ToolName)
				output += errorView("✗ ") + toolDescView(desc)
				if r.Message != "" && r.Message != "no install command available for distro: " {
					output += toolDescView(" — " + r.Message)
				}
				output += "\n"
			}
		}

		// Toggle log
		toggle := "▼ "+a.i18n.T("view_log")
		if a.model.ShowLog {
			toggle = "▲ "+a.i18n.T("hide_log")
		}
		output += "\n" + footerView(toggle+" | Enter/Space/Esc/q: Salir")

		// Log desplegable (errores)
		if a.model.ShowLog {
			output += "\n" + errorView("─── Log ───") + "\n"
			for _, r := range a.model.ProgressResults {
				if !r.Success && r.Message != "" && r.Message != "no install command available for distro: " {
					output += errorView("[") + toolNameView(r.ToolName) + errorView("]\n")
					output += toolDescView("  "+r.Message) + "\n"
				}
			}
			output += errorView("────────────────") + "\n"
		}
	} else {
		output += "\n" + footerView("Enter/Space/Esc/q: Salir")
	}

	return output
}

// getToolDescription busca la descripción de una tool en el repositorio.
func (a *App) getToolDescription(toolName string) string {
	if t := a.repo.GetByID(toolName); t != nil && t.Description != "" {
		return t.Description
	}
	return toolName
}

func (a *App) progressView() string {
	current := a.model.GetCurrentProgressTool()

	percent := 0
	if len(a.model.ProgressTools) > 0 {
		percent = (a.model.ProgressIdx * 100) / len(a.model.ProgressTools)
	}

	barWidth := 30
	filled := (barWidth * a.model.ProgressIdx) / len(a.model.ProgressTools)
	if len(a.model.ProgressTools) == 0 {
		filled = 0
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	output := headerView(a.i18n.T("installing")) + "\n\n"
	output += fmt.Sprintf("  %s %d%%\n\n", bar, percent)

	// Tool actual
	if a.model.ProgressIdx < len(a.model.ProgressTools) {
		desc := a.getToolDescription(current.Name)
		prevStatus := a.model.ProgressStatus[current.Name]
		if prevStatus {
			output += successView("  ✓ " + desc + "\n")
		} else {
			output += errorView("  ✗ " + desc + "\n")
		}
		// Output del comando
		if a.model.ProgressLastOutput != "" {
			output += toolDescView("  "+a.model.ProgressLastOutput) + "\n"
		}
	}

	// Últimos outputs (últimas 3 lines)
	var recentOutputs []string
	for i := len(a.model.ProgressResults) - 1; i >= 0 && len(recentOutputs) < 3; i-- {
		r := a.model.ProgressResults[i]
		if r.Message != "" && !strings.Contains(r.Message, "already installed") {
			recentOutputs = append(recentOutputs, r.Message)
		}
	}
	// Mostrar últimos outputs
	if len(recentOutputs) > 0 {
		output += "\n" + toolDescView("─── Output ───")
		for _, line := range recentOutputs {
			if len(line) > 60 {
				line = line[:60] + "..."
			}
			output += "\n" + toolDescView("  "+line)
		}
		output += "\n" + toolDescView("────────────────")
	}

	// Stats
	total, success, failed := a.model.GetProgressStats()
	output += "\n"
	output += fmt.Sprintf("  %s: %d | ", toolDescView(a.i18n.T("total")), total)
	output += fmt.Sprintf("%s: %d | ", successView("✓"), success)
	output += fmt.Sprintf("%s: %d\n", errorView("✗"), failed)

	if a.model.IsProgressDone() {
		output += "\n" + footerView("Enter: "+a.i18n.T("continue")+" | Esc: "+a.i18n.T("back"))
	} else {
		output += footerView(fmt.Sprintf("%s %d/%d...", a.i18n.T("installing"), a.model.ProgressIdx+1, total))
	}

	return output
}

func (a *App) settingsView() string {
	items := []string{
		"🌐 " + a.i18n.T("language") + ": " + a.model.CurrentLang,
		"🎨 " + a.i18n.T("select_category") + ": " + a.model.CurrentTheme,
		"⬅️ " + a.i18n.T("back"),
	}

	output := headerView("Vitualizz Space - Configuración") + "\n\n"
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
	output := headerView("Vitualizz Space") + "\n\n"
	output += toolDescView("github.com/vitualizz") + "\n"
	output += toolDescView("vitualizz.vercel.app") + "\n\n"
	output += toolNameView("v1.0.0") + "\n\n"
	output += footerView("Press Esc to " + a.i18n.T("back"))
	return output
}

// Output helpers
func headerView(text string) string     { return "\033[36m\033[1m" + text + "\033[0m" }
func footerView(text string) string     { return "\033[90m" + text + "\033[0m" }
func selectedView(text string) string  { return "\033[32m" + text + "\033[0m" }
func unselectedView(text string) string { return "\033[90m" + text + "\033[0m" }
func toolNameView(text string) string   { return "\033[38;5;212m" + text + "\033[0m" }
func toolDescView(text string) string  { return "\033[38;5;245m" + text + "\033[0m" }
func successView(text string) string   { return "\033[32m" + text + "\033[0m" }
func errorView(text string) string     { return "\033[31m" + text + "\033[0m" }
func warningView(text string) string    { return "\033[33m" + text + "\033[0m" }

// Messages
type statusCheckCompleteMsg struct {
	status map[string]bool
}

type progressUpdateMsg struct {
	tool     entities.Tool
	success  bool
	message  string
}

type installCompleteMsg struct {
	results []*entities.InstallResult
}

type errMsg struct{}

var _ tea.Model = (*App)(nil)