package models

import "github.com/vitualizz/vitualizz-devstack/internal/domain/entities"
import "github.com/vitualizz/vitualizz-devstack/internal/domain/interfaces"

// ViewState representa los estados de la UI.
type ViewState int

const (
	StateLanguageSelect ViewState = iota
	StateThemeSelect
	StateMainMenu
	StatePackageSelect
	StateProgress     // pantalla de progreso
	StateThanks     // después de instalar
	StateInstalling
	StateSettings
	StateAbout
)

// AppModel es el estado de la aplicación.
type AppModel struct {
	ViewState   ViewState
	Packages   []interfaces.Package
	SelectedIdx int

	// Progreso de instalación
	ProgressTools []entities.Tool
	ProgressIdx  int
	ProgressResults []entities.InstallResult
	ProgressStatus map[string]bool // tool → éxito
	ProgressLastOutput string      // último output del comando
	ShowLog       bool             // mostrar log completo

	// Status
	ToolStatus map[string]bool
	StatusChecked bool

	CurrentLang  string
	CurrentTheme string
	IsLoading   bool
	LoadingMessage string
	MainMenuChoice int
	SettingsChoice int
	ToolChoice int
	Results []entities.InstallResult
}

// NewAppModel crea un modelo inicial.
func NewAppModel() *AppModel {
	return &AppModel{
		ViewState:      StateLanguageSelect,
		SelectedIdx:    0,
		ToolStatus:    make(map[string]bool),
		ProgressStatus: make(map[string]bool),
		StatusChecked:  false,
		CurrentLang:   "es",
		CurrentTheme:  "vitualizz",
		MainMenuChoice: 0,
		SettingsChoice: 0,
		ToolChoice:    0,
		IsLoading:     false,
	}
}

// SetPackages configura los paquetes.
func (m *AppModel) SetPackages(pkgs []interfaces.Package) {
	m.Packages = pkgs
}

// GetSelectedPackage devuelve el paquete actual.
func (m *AppModel) GetSelectedPackage() *interfaces.Package {
	if m.SelectedIdx >= 0 && m.SelectedIdx < len(m.Packages) {
		return &m.Packages[m.SelectedIdx]
	}
	return nil
}

// TogglePackageSelection marca/desmarca un paquete.
func (m *AppModel) TogglePackageSelection(idx int) {
	if idx < 0 || idx >= len(m.Packages) {
		return
	}
	m.Packages[idx].Selected = !m.Packages[idx].Selected
}

// IsPackageSelected devuelve si el paquete está seleccionado.
func (m *AppModel) IsPackageSelected(idx int) bool {
	if idx < 0 || idx >= len(m.Packages) {
		return false
	}
	return m.Packages[idx].Selected
}

// SelectAll marca todos los paquetes.
func (m *AppModel) SelectAll() {
	for i := range m.Packages {
		m.Packages[i].Selected = true
	}
}

// DeselectAll desmarca todos los paquetes.
func (m *AppModel) DeselectAll() {
	for i := range m.Packages {
		m.Packages[i].Selected = false
	}
}

// GetSelectedPackages devuelve los paquetes seleccionados.
func (m *AppModel) GetSelectedPackages() []interfaces.Package {
	var selected []interfaces.Package
	for _, pkg := range m.Packages {
		if pkg.Selected {
			selected = append(selected, pkg)
		}
	}
	return selected
}

// GetSelectedTools devuelve TODAS las tools de paquetes seleccionados.
func (m *AppModel) GetSelectedTools() []entities.Tool {
	var tools []entities.Tool
	for _, pkg := range m.Packages {
		if pkg.Selected {
			tools = append(tools, pkg.Tools...)
		}
	}
	return tools
}

// StartProgress inicializa la pantalla de progreso.
func (m *AppModel) StartProgress(tools []entities.Tool) {
	m.ViewState = StateProgress
	m.ProgressTools = tools
	m.ProgressIdx = 0
	m.ProgressResults = nil
	m.ProgressStatus = make(map[string]bool)
}

// UpdateProgress actualiza el progreso de una tool.
func (m *AppModel) UpdateProgress(tool entities.Tool, success bool, msg string) {
	m.ProgressStatus[tool.Name] = success
	m.ProgressResults = append(m.ProgressResults, entities.InstallResult{
		ToolName: tool.Name,
		Success: success,
		Message: msg,
	})
	// Guardar output para mostrar en log
	m.ProgressLastOutput = msg
	m.ProgressIdx++
}

// GetCurrentProgressTool devuelve la tool que se está instalando.
func (m *AppModel) GetCurrentProgressTool() entities.Tool {
	if m.ProgressIdx < len(m.ProgressTools) {
		return m.ProgressTools[m.ProgressIdx]
	}
	return entities.Tool{}
}

// IsProgressDone devuelve si terminó la instalación.
func (m *AppModel) IsProgressDone() bool {
	return m.ProgressIdx >= len(m.ProgressTools)
}

// GetProgressStats devuelve estadísticas.
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