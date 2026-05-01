// Package views renders the TUI screens for Vitualizz DevStack.
package views

import (
	"fmt"
	"strings"

	"github.com/vitualizz/vitualizz-devstack/internal/domain/interfaces"
	"github.com/vitualizz/vitualizz-devstack/internal/ui/models"
)

// Renderer holds the data needed to render all screens.
type Renderer struct {
	Model *models.AppModel
	Repo  interfaces.ToolRepository
	I18n  I18nReader
}

// I18nReader is the subset of i18n needed for rendering.
type I18nReader interface {
	T(key string) string
}

// View dispatches to the correct screen based on model state.
func (r *Renderer) View() string {
	switch r.Model.ViewState {
	case models.StateLanguageSelect:
		return r.LanguageSelectView()
	case models.StateThemeSelect:
		return r.ThemeSelectView()
	case models.StateMainMenu:
		return r.MainMenuView()
	case models.StateThanks:
		return r.ThanksView()
	case models.StateProgress:
		return r.ProgressView()
	case models.StateSettings:
		return r.SettingsView()
	case models.StateAbout:
		return r.AboutView()
	}
	return ""
}

// =============================================================================
// Style helpers
// =============================================================================

func Header(text string) string     { return "\033[36m\033[1m" + text + "\033[0m" }
func Footer(text string) string     { return "\033[90m" + text + "\033[0m" }
func Selected(text string) string   { return "\033[32m" + text + "\033[0m" }
func Unselected(text string) string { return "\033[90m" + text + "\033[0m" }
func ToolName(text string) string   { return "\033[38;5;212m" + text + "\033[0m" }
func ToolDesc(text string) string   { return "\033[38;5;245m" + text + "\033[0m" }
func Success(text string) string    { return "\033[32m" + text + "\033[0m" }
func Error(text string) string      { return "\033[31m" + text + "\033[0m" }
func Warning(text string) string    { return "\033[33m" + text + "\033[0m" }

// =============================================================================
// Language Select
// =============================================================================

func (r *Renderer) LanguageSelectView() string {
	languages := []struct {
		code string
		name string
	}{
		{"es", "Español"},
		{"en", "English"},
	}

	output := Header("Vitualizz DevStack - Idioma") + "\n\n"
	for i, lang := range languages {
		if i == r.Model.SettingsChoice {
			output += Selected("► " + lang.name) + "\n"
		} else {
			output += Unselected("  " + lang.name) + "\n"
		}
	}
	output += "\n" + Footer("↑↓: " + r.I18n.T("select_option") + " | Enter: " + r.I18n.T("confirm"))
	return output
}

// =============================================================================
// Theme Select
// =============================================================================

func (r *Renderer) ThemeSelectView() string {
	themes := r.Repo.GetThemes()
	output := Header("Vitualizz DevStack - Themes") + "\n\n"

	if len(themes) == 0 {
		output += Warning(r.I18n.T("no_tools_found"))
		return output
	}

	for i, theme := range themes {
		if i == r.Model.ToolChoice {
			output += Selected("► " + theme.DisplayName) + "\n"
		} else {
			output += Unselected("  " + theme.DisplayName) + "\n"
		}
	}
	output += "\n" + Footer("↑↓: " + r.I18n.T("select_option") + " | Enter: " + r.I18n.T("confirm") + " | Esc: " + r.I18n.T("back"))
	return output
}

// =============================================================================
// Main Menu
// =============================================================================

func (r *Renderer) MainMenuView() string {
	items := []string{
		"⚡ " + r.I18n.T("install"),
		"🗑️ " + r.I18n.T("uninstall"),
		"⚙️ " + r.I18n.T("settings"),
		"ℹ️ " + r.I18n.T("about"),
		"🚪 " + r.I18n.T("exit"),
	}

	output := asciiBanner() + "\n"
	if r.Model.InDocker {
		output += Warning("  🐳 Docker mode — display tools (kitty, docker) are skipped") + "\n"
	}
	output += "\n"
	for i, item := range items {
		if i == r.Model.MainMenuChoice {
			output += Selected("► " + item) + "\n"
		} else {
			output += Unselected("  " + item) + "\n"
		}
	}
	output += "\n" + Footer("↑↓: " + r.I18n.T("select_option") + " | Enter: " + r.I18n.T("confirm"))
	return output
}

// =============================================================================
// Thanks (post-install/uninstall)
// =============================================================================

func (r *Renderer) ThanksView() string {
	total, success, failed := r.Model.GetProgressStats()

	if r.Model.IsUninstallMode {
		return r.thanksUninstallView(total, success, failed)
	}
	return r.thanksInstallView(total, success, failed)
}

// thanksInstallView shows a clean dashboard-style post-install screen.
func (r *Renderer) thanksInstallView(total, success, failed int) string {
	output := asciiBannerSmall() + "\n\n"
	output += Selected("✓ DevStack installed successfully") + "\n\n"

	// Stats bar
	bar := r.statsBar(total, success, failed)
	output += ToolDesc(bar) + "\n\n"

	// If all good, short message
	if failed == 0 {
		output += Success("  All " + fmt.Sprint(total) + " tools installed without issues.") + "\n\n"
	} else {
		// Show only failures
		output += Error("✗ " + r.I18n.T("some_errors")) + "\n\n"
		for _, res := range r.Model.ProgressResults {
			if !res.Success {
				desc := r.toolDescription(res.ToolName)
				output += Error("  ✗ ") + ToolDesc(desc)
				if res.Message != "" && res.Message != "no install command available for distro: " {
					output += ToolDesc(" — " + r.truncate(res.Message, 80))
				}
				output += "\n"
			}
		}

		toggle := "▼ " + r.I18n.T("view_log")
		if r.Model.ShowLog {
			toggle = "▲ " + r.I18n.T("hide_log")
		}
		output += "\n" + Footer(toggle + " | Enter/Space/Esc/q: Salir")

		if r.Model.ShowLog {
			output += "\n" + Error("─── Full Log ───") + "\n"
			for _, res := range r.Model.ProgressResults {
				if !res.Success && res.Message != "" {
					output += Error("[") + ToolName(res.ToolName) + Error("]\n")
					output += ToolDesc("  " + res.Message) + "\n"
				}
			}
			output += Error("────────────────") + "\n"
		}
	}

	output += "\n" + Footer("Con amor @vitualizz")
	return output
}

// thanksUninstallView shows a clean post-uninstall screen.
func (r *Renderer) thanksUninstallView(total, success, failed int) string {
	output := Header("🗑️ Uninstall Complete") + "\n\n"

	bar := r.statsBar(total, success, failed)
	output += ToolDesc(bar) + "\n\n"

	if failed == 0 {
		output += Success("  All " + fmt.Sprint(total) + " tools removed.") + "\n\n"
	} else {
		output += Error("✗ Some tools could not be removed:") + "\n\n"
		for _, res := range r.Model.ProgressResults {
			if !res.Success {
				desc := r.toolDescription(res.ToolName)
				output += Error("  ✗ ") + ToolDesc(desc) + "\n"
			}
		}
	}

	output += Footer("Enter/Space/Esc/q: Salir")
	return output
}

// statsBar builds a visual stats line like "Total: 45  |  ✓ 37  |  ✗ 5  |  Skipped: 3"
func (r *Renderer) statsBar(total, success, failed int) string {
	skipped := total - success - failed
	bar := fmt.Sprintf("Total: %d  |  %s %d  |  %s %d",
		total,
		Success("✓"), success,
		Error("✗"), failed,
	)
	if skipped > 0 {
		bar += fmt.Sprintf("  |  %s %d", ToolDesc("⊘"), skipped)
	}
	return bar
}

// =============================================================================
// Progress (install/uninstall)
// =============================================================================

func (r *Renderer) ProgressView() string {
	current := r.Model.GetCurrentProgressTool()

	percent := 0
	if len(r.Model.ProgressTools) > 0 {
		percent = (r.Model.ProgressIdx * 100) / len(r.Model.ProgressTools)
	}

	barWidth := 30
	filled := 0
	if len(r.Model.ProgressTools) > 0 {
		filled = (barWidth * r.Model.ProgressIdx) / len(r.Model.ProgressTools)
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	action := r.I18n.T("installing")
	if r.Model.IsUninstallMode {
		action = r.I18n.T("uninstalling")
	}

	output := Header(action) + "\n\n"
	output += fmt.Sprintf("  %s %d%%\n\n", bar, percent)

	// Current tool
	if r.Model.ProgressIdx < len(r.Model.ProgressTools) {
		desc := r.toolDescription(current.Name)
		prevStatus := r.Model.ProgressStatus[current.Name]
		if prevStatus {
			output += Success("  ✓ " + desc) + "\n"
		} else {
			output += Error("  ✗ " + desc) + "\n"
		}
		if r.Model.ProgressLastOutput != "" {
			output += ToolDesc("  " + r.Model.ProgressLastOutput) + "\n"
		}
	}

	// Recent outputs (last 3)
	var recentOutputs []string
	for i := len(r.Model.ProgressResults) - 1; i >= 0 && len(recentOutputs) < 3; i-- {
		res := r.Model.ProgressResults[i]
		if res.Message != "" && !strings.Contains(res.Message, "already installed") && !strings.Contains(res.Message, "not installed") {
			recentOutputs = append(recentOutputs, res.Message)
		}
	}
	if len(recentOutputs) > 0 {
		output += "\n" + ToolDesc("─── Output ───")
		for _, line := range recentOutputs {
			if len(line) > 60 {
				line = line[:60] + "..."
			}
			output += "\n" + ToolDesc("  " + line)
		}
		output += "\n" + ToolDesc("────────────────")
	}

	// Stats
	total, successCount, failed := r.Model.GetProgressStats()
	output += "\n"
	output += fmt.Sprintf("  %s: %d | ", ToolDesc(r.I18n.T("total")), total)
	output += fmt.Sprintf("%s: %d | ", Success("✓"), successCount)
	output += fmt.Sprintf("%s: %d\n", Error("✗"), failed)

	if r.Model.IsProgressDone() {
		output += "\n" + Footer("Enter: " + r.I18n.T("continue") + " | Esc: " + r.I18n.T("back"))
	} else {
		output += Footer(fmt.Sprintf("%s %d/%d...", action, r.Model.ProgressIdx+1, total))
	}

	return output
}

// =============================================================================
// Settings
// =============================================================================

func (r *Renderer) SettingsView() string {
	items := []string{
		"🌐 " + r.I18n.T("language") + ": " + r.Model.CurrentLang,
		"🎨 " + r.I18n.T("select_category") + ": " + r.Model.CurrentTheme,
		"⬅️ " + r.I18n.T("back"),
	}

	output := Header("Vitualizz DevStack - Configuración") + "\n\n"
	for i, item := range items {
		if i == r.Model.SettingsChoice {
			output += Selected("► " + item) + "\n"
		} else {
			output += Unselected("  " + item) + "\n"
		}
	}
	output += "\n" + Footer("↑↓: " + r.I18n.T("select_option") + " | Enter: " + r.I18n.T("confirm"))
	return output
}

// =============================================================================
// About
// =============================================================================

func (r *Renderer) AboutView() string {
	output := Header("Vitualizz DevStack") + "\n\n"
	output += ToolDesc("github.com/vitualizz") + "\n"
	output += ToolDesc("vitualizz.vercel.app") + "\n\n"
	output += ToolName("v1.0.0") + "\n\n"
	output += Footer("Press Esc to " + r.I18n.T("back"))
	return output
}

// =============================================================================
// Helpers
// =============================================================================

func (r *Renderer) toolDescription(name string) string {
	if t := r.Repo.GetByID(name); t != nil && t.Description != "" {
		return t.Description
	}
	return name
}

// truncate shortens a string to maxLen, adding "..." if needed.
func (r *Renderer) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// MenuItemCount returns the number of items in the main menu.
func (r *Renderer) MenuItemCount() int { return 5 }

// SettingsItemCount returns the number of items in settings.
func (r *Renderer) SettingsItemCount() int { return 3 }

// =============================================================================
// ASCII Art Banners
// =============================================================================

// asciiBanner returns the full banner for the main menu.
func asciiBanner() string {
	lines := []string{
		"\033[36m\033[1m" + " █▀▀ █░█ █▀▀ █░░ █▀▀ ▀█▀" + "\033[0m",
		"\033[36m\033[1m" + " █░░ █▀█ █▀▀ █▄▄ █▄░ ░█░" + "\033[0m",
		"\033[36m\033[1m" + " ▀▀▀ ▀░▀ ▀▀▀ ▀▀▀ ▀░▀ ░▀▀" + "\033[0m",
		"",
		"\033[36m\033[1m" + "  ▀█▀ █▀█ █▀▀ █▀█" + "\033[0m",
		"\033[36m\033[1m" + "  ░█░ █▀▀ █▄▄ █▀▄" + "\033[0m",
		"\033[36m\033[1m" + "  ▀▀▀ ▀░░ ▀▀▀ ▀░▀" + "\033[0m",
		"",
		"\033[36m\033[1m" + "  D e v S t a c k" + "\033[0m",
	}
	return strings.Join(lines, "\n")
}

// asciiBannerSmall returns a compact banner for the thanks screen.
func asciiBannerSmall() string {
	lines := []string{
		"\033[36m\033[1m" + " █▀▀ █░█ █▀▀ █░░ █▀▀ ▀█▀  Dev" + "\033[0m",
		"\033[36m\033[1m" + " █░░ █▀█ █▀▀ █▄▄ █▄░ ░█░  Stack" + "\033[0m",
		"\033[36m\033[1m" + " ▀▀▀ ▀░▀ ▀▀▀ ▀▀▀ ▀░▀ ░▀▀" + "\033[0m",
	}
	return strings.Join(lines, "\n")
}
