package usecases

import (
	"github.com/vitualizz/envsetup/internal/domain/entities"
	"github.com/vitualizz/envsetup/internal/domain/interfaces"
)

// InstallToolUseCase installs a single tool.
type InstallToolUseCase struct {
	installer interfaces.InstallerPort
	repo     interfaces.ToolRepository
}

func NewInstallToolUseCase(installer interfaces.InstallerPort, repo interfaces.ToolRepository) *InstallToolUseCase {
	return &InstallToolUseCase{installer: installer, repo: repo}
}

func (uc *InstallToolUseCase) Execute(tool *entities.Tool) (*entities.InstallResult, error) {
	installed, err := uc.installer.IsInstalled(tool)
	if err != nil {
		return nil, err
	}

	if installed {
		return &entities.InstallResult{
			ToolName: tool.Name,
			Success: true,
			Message: "already installed",
		}, nil
	}

	return uc.installer.Install(tool)
}

// UninstallToolUseCase uninstalls a single tool.
type UninstallToolUseCase struct {
	installer interfaces.InstallerPort
	repo     interfaces.ToolRepository
}

func NewUninstallToolUseCase(installer interfaces.InstallerPort, repo interfaces.ToolRepository) *UninstallToolUseCase {
	return &UninstallToolUseCase{installer: installer, repo: repo}
}

func (uc *UninstallToolUseCase) Execute(tool *entities.Tool) (*entities.InstallResult, error) {
	installed, err := uc.installer.IsInstalled(tool)
	if err != nil {
		return nil, err
	}

	if !installed {
		return &entities.InstallResult{
			ToolName: tool.Name,
			Success: true,
			Message: "not installed",
		}, nil
	}

	return uc.installer.Uninstall(tool)
}

// ListToolsUseCase lists available tools.
type ListToolsUseCase struct {
	repo interfaces.ToolRepository
}

func NewListToolsUseCase(repo interfaces.ToolRepository) *ListToolsUseCase {
	return &ListToolsUseCase{repo: repo}
}

func (uc *ListToolsUseCase) Execute() []entities.Tool {
	return uc.repo.GetAll()
}

func (uc *ListToolsUseCase) GetMainTools() []entities.Tool {
	return uc.repo.GetMainTools()
}

func (uc *ListToolsUseCase) ExecuteByCategory(category entities.Category) []entities.Tool {
	return uc.repo.GetByCategory(category)
}

// CheckInstallationUseCase checks if a tool is installed.
type CheckInstallationUseCase struct {
	installer interfaces.InstallerPort
}

func NewCheckInstallationUseCase(installer interfaces.InstallerPort) *CheckInstallationUseCase {
	return &CheckInstallationUseCase{installer: installer}
}

func (uc *CheckInstallationUseCase) Execute(tool *entities.Tool) (bool, error) {
	return uc.installer.IsInstalled(tool)
}

// BatchCheckStatusUseCase checks installation status for multiple tools.
type BatchCheckStatusUseCase struct {
	installer interfaces.InstallerPort
}

func NewBatchCheckStatusUseCase(installer interfaces.InstallerPort) *BatchCheckStatusUseCase {
	return &BatchCheckStatusUseCase{installer: installer}
}

func (uc *BatchCheckStatusUseCase) Execute(tools []entities.Tool) map[string]bool {
	status := make(map[string]bool, len(tools))
	for _, tool := range tools {
		installed, _ := uc.installer.IsInstalled(&tool)
		status[tool.Name] = installed
	}
	return status
}

// BatchInstallUseCase installs multiple tools with dependency resolution.
type BatchInstallUseCase struct {
	installer interfaces.InstallerPort
	repo     interfaces.ToolRepository
}

func NewBatchInstallUseCase(installer interfaces.InstallerPort, repo interfaces.ToolRepository) *BatchInstallUseCase {
	return &BatchInstallUseCase{installer: installer, repo: repo}
}

// Execute installs tools in dependency order.
// Dependencies are installed first if not already present.
func (uc *BatchInstallUseCase) Execute(tools []*entities.Tool) []*entities.InstallResult {
	results := make([]*entities.InstallResult, 0, len(tools)*2)

	for _, tool := range tools {
		res := uc.installWithDeps(tool, results)
		results = append(results, res)
	}

	return results
}

// installWithDeps installs a tool and its dependencies.
func (uc *BatchInstallUseCase) installWithDeps(tool *entities.Tool, done []*entities.InstallResult) *entities.InstallResult {
	// Skip bundles (no install command)
	if !tool.HasInstallCommand() {
		return &entities.InstallResult{
			ToolName: tool.Name,
			Success: true,
			Message: "bundle (no install command)",
		}
	}

	// Check if already done or installed
	installed, _ := uc.installer.IsInstalled(tool)
	if installed {
		for _, r := range done {
			if r.ToolName == tool.Name && r.Success {
				return &entities.InstallResult{
					ToolName: tool.Name,
					Success: true,
					Message: "already installed",
				}
			}
		}
	}

	// Install dependencies first
	for _, depName := range tool.DependsOn {
		dep := uc.repo.GetByID(depName)
		if dep == nil || !dep.HasInstallCommand() {
			continue
		}
		if alreadyDone(dep.Name, done) {
			continue
		}

		depInstalled, _ := uc.installer.IsInstalled(dep)
		if !depInstalled {
			uc.installer.Install(dep)
		}
	}

	// Install the tool
	result, err := uc.installer.Install(tool)
	if err != nil {
		return &entities.InstallResult{
			ToolName: tool.Name,
			Success: false,
			Message: err.Error(),
		}
	}

	return result
}

// alreadyDone checks if a tool was already successfully installed in this batch.
func alreadyDone(name string, done []*entities.InstallResult) bool {
	for _, r := range done {
		if r.ToolName == name && r.Success {
			return true
		}
	}
	return false
}