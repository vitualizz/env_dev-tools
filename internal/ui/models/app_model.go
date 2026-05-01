package models

import "github.com/vitualizz/vitualizz-devstack/internal/domain/entities"

// ViewState represents the TUI screens.
type ViewState int

const (
	StateLanguageSelect ViewState = iota
	StateThemeSelect
	StateMainMenu
	StateProgress // install/uninstall progress
	StateThanks   // post-install/uninstall summary
	StateSettings
	StateAbout
)

// AppModel is the application state.
type AppModel struct {
	ViewState    ViewState
	CurrentLang  string
	CurrentTheme string
	IsLoading    bool
	MainMenuChoice int
	SettingsChoice int
	ToolChoice   int
	InDocker     bool // set to true when running inside a Docker container

	// Progress state (shared between install and uninstall)
	ProgressTools   []entities.Tool
	ProgressIdx     int
	ProgressResults []entities.InstallResult
	ProgressStatus  map[string]bool // tool name → success
	ProgressLastOutput string
	IsUninstallMode bool
	ShowLog        bool
	Results        []entities.InstallResult
	LogPath        string // path to install log file
}

// NewAppModel creates an initial model.
func NewAppModel() *AppModel {
	return &AppModel{
		ViewState:      StateLanguageSelect,
		CurrentLang:    "es",
		CurrentTheme:   "vitualizz",
		MainMenuChoice: 0,
		SettingsChoice: 0,
		ToolChoice:     0,
		ProgressStatus: make(map[string]bool),
	}
}

// StartProgress initializes the progress screen with the given tools.
func (m *AppModel) StartProgress(tools []entities.Tool) {
	m.ViewState = StateProgress
	m.ProgressTools = tools
	m.ProgressIdx = 0
	m.ProgressResults = nil
	m.ProgressStatus = make(map[string]bool)
	m.IsUninstallMode = false
}

// StartUninstallProgress initializes the progress screen for uninstall.
func (m *AppModel) StartUninstallProgress(tools []entities.Tool) {
	m.ViewState = StateProgress
	m.ProgressTools = tools
	m.ProgressIdx = 0
	m.ProgressResults = nil
	m.ProgressStatus = make(map[string]bool)
	m.IsUninstallMode = true
}

// UpdateProgress records the result of installing/uninstalling a tool.
func (m *AppModel) UpdateProgress(tool entities.Tool, success bool, msg string) {
	m.ProgressStatus[tool.Name] = success
	m.ProgressResults = append(m.ProgressResults, entities.InstallResult{
		ToolName: tool.Name,
		Success:  success,
		Message:  msg,
	})
	m.ProgressLastOutput = msg
	m.ProgressIdx++
}

// GetCurrentProgressTool returns the tool being processed.
func (m *AppModel) GetCurrentProgressTool() entities.Tool {
	if m.ProgressIdx < len(m.ProgressTools) {
		return m.ProgressTools[m.ProgressIdx]
	}
	return entities.Tool{}
}

// IsProgressDone returns true if all tools have been processed.
func (m *AppModel) IsProgressDone() bool {
	return m.ProgressIdx >= len(m.ProgressTools)
}

// GetProgressStats returns total, success, and failed counts.
func (m *AppModel) GetProgressStats() (total, success, failed int) {
	total = len(m.ProgressTools)
	for _, r := range m.ProgressResults {
		if r.Success {
			success++
		} else {
			failed++
		}
	}
	return total, success, failed
}
