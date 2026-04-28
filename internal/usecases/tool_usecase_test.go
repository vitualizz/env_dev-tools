package usecases_test

import (
	"errors"
	"testing"

	"github.com/vitualizz/envsetup/internal/domain/entities"
	"github.com/vitualizz/envsetup/internal/domain/interfaces"
	"github.com/vitualizz/envsetup/internal/usecases"
)

// --- mockInstaller ---

type mockInstaller struct {
	installed map[string]bool
	failOn    map[string]error
	calls     []string
}

func newMockInstaller() *mockInstaller {
	return &mockInstaller{
		installed: make(map[string]bool),
		failOn:    make(map[string]error),
	}
}

func (m *mockInstaller) Install(tool *entities.Tool) (*entities.InstallResult, error) {
	m.calls = append(m.calls, "install:"+tool.Name)
	if err, ok := m.failOn[tool.Name]; ok {
		return &entities.InstallResult{ToolName: tool.Name, Success: false}, err
	}
	m.installed[tool.Name] = true
	return &entities.InstallResult{ToolName: tool.Name, Success: true, Message: "ok"}, nil
}

func (m *mockInstaller) Uninstall(tool *entities.Tool) (*entities.InstallResult, error) {
	m.calls = append(m.calls, "uninstall:"+tool.Name)
	delete(m.installed, tool.Name)
	return &entities.InstallResult{ToolName: tool.Name, Success: true}, nil
}

func (m *mockInstaller) IsInstalled(tool *entities.Tool) (bool, error) {
	return m.installed[tool.Name], nil
}

var _ interfaces.InstallerPort = (*mockInstaller)(nil)

// --- mockRepo ---

type mockRepo struct {
	tools []entities.Tool
}

func (r *mockRepo) GetAll() []entities.Tool                               { return r.tools }
func (r *mockRepo) GetMainTools() []entities.Tool                         { return r.tools }
func (r *mockRepo) GetByCategory(_ entities.Category) []entities.Tool     { return nil }
func (r *mockRepo) GetDependencies(_ string) []entities.Tool               { return nil }
func (r *mockRepo) GetDependents(_ string) []entities.Tool                 { return nil }
func (r *mockRepo) GetPackages() []interfaces.Package                      { return nil }
func (r *mockRepo) GetThemes() []entities.Theme                            { return nil }
func (r *mockRepo) GetThemeByName(_ string) *entities.Theme                { return nil }
func (r *mockRepo) Save(_ entities.Tool)                                   {}
func (r *mockRepo) GetByID(name string) *entities.Tool {
	for i := range r.tools {
		if r.tools[i].Name == name {
			return &r.tools[i]
		}
	}
	return nil
}

var _ interfaces.ToolRepository = (*mockRepo)(nil)

// --- InstallToolUseCase ---

func TestInstallToolUseCase_AlreadyInstalled(t *testing.T) {
	installer := newMockInstaller()
	installer.installed["git"] = true

	uc := usecases.NewInstallToolUseCase(installer, &mockRepo{})
	tool := &entities.Tool{Name: "git", Install: map[entities.Distro]string{entities.DistroAll: "apt install git"}}

	result, err := uc.Execute(tool)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Message != "already installed" {
		t.Errorf("result.Message = %q, want %q", result.Message, "already installed")
	}
	for _, c := range installer.calls {
		if c == "install:git" {
			t.Error("Install() was called even though tool was already installed")
		}
	}
}

func TestInstallToolUseCase_NotInstalled(t *testing.T) {
	installer := newMockInstaller()
	uc := usecases.NewInstallToolUseCase(installer, &mockRepo{})
	tool := &entities.Tool{Name: "git", Install: map[entities.Distro]string{entities.DistroAll: "apt install git"}}

	result, err := uc.Execute(tool)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Success {
		t.Errorf("result.Success = false, want true")
	}
}

func TestInstallToolUseCase_InstallFails(t *testing.T) {
	installer := newMockInstaller()
	installer.failOn["broken"] = errors.New("network timeout")
	uc := usecases.NewInstallToolUseCase(installer, &mockRepo{})
	tool := &entities.Tool{Name: "broken", Install: map[entities.Distro]string{entities.DistroAll: "exit 1"}}

	_, err := uc.Execute(tool)
	if err == nil {
		t.Error("Execute() error = nil, want error")
	}
}

// --- UninstallToolUseCase ---

func TestUninstallToolUseCase_NotInstalled(t *testing.T) {
	installer := newMockInstaller()
	uc := usecases.NewUninstallToolUseCase(installer, &mockRepo{})
	tool := &entities.Tool{Name: "git"}

	result, err := uc.Execute(tool)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Message != "not installed" {
		t.Errorf("result.Message = %q, want %q", result.Message, "not installed")
	}
	for _, c := range installer.calls {
		if c == "uninstall:git" {
			t.Error("Uninstall() was called even though tool was not installed")
		}
	}
}

func TestUninstallToolUseCase_Installed(t *testing.T) {
	installer := newMockInstaller()
	installer.installed["git"] = true
	uc := usecases.NewUninstallToolUseCase(installer, &mockRepo{})
	tool := &entities.Tool{Name: "git"}

	result, err := uc.Execute(tool)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Success {
		t.Errorf("result.Success = false, want true")
	}
	if installer.installed["git"] {
		t.Error("tool still marked as installed after Uninstall()")
	}
}

// --- BatchCheckStatusUseCase ---

func TestBatchCheckStatusUseCase(t *testing.T) {
	installer := newMockInstaller()
	installer.installed["git"] = true
	installer.installed["nvim"] = true

	uc := usecases.NewBatchCheckStatusUseCase(installer)
	tools := []entities.Tool{
		{Name: "git"},
		{Name: "nvim"},
		{Name: "missing"},
	}

	status := uc.Execute(tools)

	if !status["git"] {
		t.Error("git should be installed")
	}
	if !status["nvim"] {
		t.Error("nvim should be installed")
	}
	if status["missing"] {
		t.Error("missing should not be installed")
	}
}

// --- BatchInstallUseCase ---

func TestBatchInstallUseCase_InstallsDependenciesFirst(t *testing.T) {
	installer := newMockInstaller()

	toolA := entities.Tool{
		Name:    "tool-a",
		Install: map[entities.Distro]string{entities.DistroAll: "echo a"},
	}
	toolB := entities.Tool{
		Name:      "tool-b",
		DependsOn: []string{"tool-a"},
		Install:   map[entities.Distro]string{entities.DistroAll: "echo b"},
	}

	repo := &mockRepo{tools: []entities.Tool{toolA, toolB}}
	uc := usecases.NewBatchInstallUseCase(installer, repo)
	uc.Execute([]*entities.Tool{&toolB})

	if !installer.installed["tool-a"] {
		t.Error("dependency tool-a was not installed before tool-b")
	}
	if !installer.installed["tool-b"] {
		t.Error("tool-b was not installed")
	}
}

func TestBatchInstallUseCase_SkipsAlreadyInstalled(t *testing.T) {
	installer := newMockInstaller()
	installer.installed["tool-a"] = true

	toolA := entities.Tool{
		Name:    "tool-a",
		Install: map[entities.Distro]string{entities.DistroAll: "echo a"},
	}
	repo := &mockRepo{tools: []entities.Tool{toolA}}
	uc := usecases.NewBatchInstallUseCase(installer, repo)
	uc.Execute([]*entities.Tool{&toolA})

	for _, c := range installer.calls {
		if c == "install:tool-a" {
			t.Error("tool-a was installed again even though already installed")
		}
	}
}

func TestBatchInstallUseCase_NoInstallCommand(t *testing.T) {
	installer := newMockInstaller()
	bundle := entities.Tool{Name: "bundle", Install: map[entities.Distro]string{}}
	repo := &mockRepo{tools: []entities.Tool{bundle}}
	uc := usecases.NewBatchInstallUseCase(installer, repo)

	results := uc.Execute([]*entities.Tool{&bundle})
	if len(results) == 0 {
		t.Fatal("Execute() returned no results")
	}
	if results[0].Message != "bundle (no install command)" {
		t.Errorf("result.Message = %q, want %q", results[0].Message, "bundle (no install command)")
	}
}
