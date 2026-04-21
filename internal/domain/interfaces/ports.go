package interfaces

import "github.com/vitualizz/envsetup/internal/domain/entities"

// ToolRepository queries tools and themes.
// All methods that can fail return an error. Read-only methods return nil.
type ToolRepository interface {
	// Tools
	GetAll() []entities.Tool
	GetMainTools() []entities.Tool
	GetByCategory(category entities.Category) []entities.Tool
	GetByID(name string) *entities.Tool
	GetDependencies(name string) []entities.Tool
	GetDependents(name string) []entities.Tool

	// Themes
	GetThemes() []entities.Theme
	GetThemeByName(name string) *entities.Theme

	// Write
	Save(tool entities.Tool)
}

// InstallerPort installs and uninstalls tools.
type InstallerPort interface {
	Install(tool *entities.Tool) (*entities.InstallResult, error)
	Uninstall(tool *entities.Tool) (*entities.InstallResult, error)
	IsInstalled(tool *entities.Tool) (bool, error)
}

// ExecutorPort executes shell commands.
type ExecutorPort interface {
	Execute(cmd string) (string, error)
	ExecuteWithOutput(cmd string) (string, error)
}

// I18nPort provides translations.
type I18nPort interface {
	Get(key string, lang string) string
	SetLanguage(lang string)
	GetAvailableLanguages() []string
}