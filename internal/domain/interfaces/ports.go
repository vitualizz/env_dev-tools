package interfaces

import "github.com/vitualizz/envsetup/internal/domain/entities"

// ToolRepository queries tools and themes.
type ToolRepository interface {
	GetAll() []entities.Tool
	GetMainTools() []entities.Tool
	GetByCategory(category entities.Category) []entities.Tool
	GetByID(name string) *entities.Tool
	GetDependencies(name string) []entities.Tool
	GetDependents(name string) []entities.Tool
	GetPackages() []Package
	GetThemes() []entities.Theme
	GetThemeByName(name string) *entities.Theme
	Save(tool entities.Tool)
}

// Package representa un paquete instalable.
type Package struct {
	Name             string
	Label            string
	Icon             string
	Description      string // una línea descriptiva
	Tools           []entities.Tool
	DefaultSelected bool   // se seleciona por defecto
	Selected        bool
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