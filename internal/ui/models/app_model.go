package models

import (
	"github.com/vitualizz/envsetup/internal/domain/entities"
)

type ViewState int

const (
	StateLanguageSelect ViewState = iota
	StateThemeSelect
	StateMainMenu
	StateToolList
	StateInstalling
	StateSettings
	StateAbout
)

type AppModel struct {
	ViewState       ViewState
	Tools           []entities.Tool
	SelectedTools   map[string]bool
	ToolStatus      map[string]bool
	StatusChecked   bool
	CurrentCategory entities.Category
	Results         []entities.InstallResult
	CurrentLang     string
	CurrentTheme    string
	IsLoading       bool
	LoadingMessage  string
	MainMenuChoice  int
	ToolChoice      int
	SettingsChoice  int
	FilterQuery     string
}

func NewAppModel() *AppModel {
	return &AppModel{
		ViewState:       StateLanguageSelect,
		SelectedTools:   make(map[string]bool),
		ToolStatus:      make(map[string]bool),
		StatusChecked:   false,
		CurrentLang:     "es",
		CurrentTheme:    "",
		MainMenuChoice:  0,
		ToolChoice:      0,
		SettingsChoice:  0,
		FilterQuery:     "",
		IsLoading:       false,
	}
}

func (m *AppModel) SetTools(tools []entities.Tool) {
	m.Tools = tools
}

func (m *AppModel) ToggleToolSelection(name string) {
	if m.SelectedTools[name] {
		delete(m.SelectedTools, name)
	} else {
		m.SelectedTools[name] = true
	}
}

func (m *AppModel) IsToolSelected(name string) bool {
	return m.SelectedTools[name]
}

func (m *AppModel) GetSelectedToolsList() []entities.Tool {
	var selected []entities.Tool
	for i := range m.Tools {
		if m.SelectedTools[m.Tools[i].Name] {
			selected = append(selected, m.Tools[i])
		}
	}
	return selected
}

func (m *AppModel) GetFilteredTools() []entities.Tool {
	if m.FilterQuery == "" {
		return m.Tools
	}
	
	var filtered []entities.Tool
	query := m.FilterQuery
	for _, tool := range m.Tools {
		if containsIgnoreCase(tool.Name, query) || containsIgnoreCase(tool.Description, query) {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}