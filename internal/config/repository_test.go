package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vitualizz/envsetup/internal/config"
	"github.com/vitualizz/envsetup/internal/domain/entities"
)

const testYAML = `
themes:
  - name: test-theme
    display_name: "Test Theme"
    category: theme
    colors:
      background: "0x1A1B26"
      foreground: "0xC0CAF5"
      normal:
        black: "0x414868"
        red: "0xF7768E"
        green: "0x73DACA"
        yellow: "0xE0AF68"
        blue: "0x7AA2F7"
        magenta: "0xBB9AF7"
        cyan: "0x7DCFFF"
        white: "0xC0CAF5"
      bright:
        black: "0x414868"
        red: "0xF7768E"
        green: "0x73DACA"
        yellow: "0xE0AF68"
        blue: "0x7AA2F7"
        magenta: "0xBB9AF7"
        cyan: "0x7DCFFF"
        white: "0xC0CAF5"

tools:
  - name: tool-a
    category: tools
    description: "Tool A"
    install:
      all: echo install-a
    uninstall:
      all: echo uninstall-a
    check: echo check-a
    enabled: true
    required: true

  - name: tool-b
    category: tools
    description: "Tool B"
    depends_on:
      - tool-a
    install:
      debian: apt-get install -y tool-b
      arch: pacman -S tool-b
    enabled: true
    required: false

  - name: tool-c
    category: shell
    description: "Tool C"
    depends_on:
      - tool-b
    install:
      all: echo install-c
    enabled: false
    required: false
`

func newTestRepo(t *testing.T) *config.ToolRepository {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tools.yaml")
	if err := os.WriteFile(path, []byte(testYAML), 0644); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}
	repo, err := config.NewToolRepository(path)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	return repo
}

func TestNewToolRepository_ValidYAML(t *testing.T) {
	repo := newTestRepo(t)
	tools := repo.GetAll()
	if len(tools) != 3 {
		t.Errorf("GetAll() returned %d tools, want 3", len(tools))
	}
}

func TestNewToolRepository_InvalidPath(t *testing.T) {
	_, err := config.NewToolRepository("/nonexistent/path/tools.yaml")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestNewToolRepository_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tools.yaml")
	if err := os.WriteFile(path, []byte("this: is: invalid: yaml: {{{"), 0644); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}

	_, err := config.NewToolRepository(path)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestGetByID(t *testing.T) {
	repo := newTestRepo(t)

	tests := []struct {
		name      string
		id        string
		wantFound bool
	}{
		{"existing tool", "tool-a", true},
		{"another existing tool", "tool-b", true},
		{"nonexistent tool", "does-not-exist", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.GetByID(tt.id)
			if tt.wantFound && got == nil {
				t.Errorf("GetByID(%q) = nil, want tool", tt.id)
			}
			if !tt.wantFound && got != nil {
				t.Errorf("GetByID(%q) = %v, want nil", tt.id, got.Name)
			}
			if got != nil && got.Name != tt.id {
				t.Errorf("GetByID(%q).Name = %q, want %q", tt.id, got.Name, tt.id)
			}
		})
	}
}

func TestGetByCategory(t *testing.T) {
	repo := newTestRepo(t)

	tests := []struct {
		name     string
		category entities.Category
		wantLen  int
	}{
		{"tools category", entities.CategoryTools, 2},
		{"shell category", entities.CategoryShell, 1},
		{"terminal category (empty)", entities.CategoryTerminal, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.GetByCategory(tt.category)
			if len(got) != tt.wantLen {
				t.Errorf("GetByCategory(%q) returned %d tools, want %d", tt.category, len(got), tt.wantLen)
			}
		})
	}
}

func TestGetMainTools(t *testing.T) {
	repo := newTestRepo(t)

	// tool-a is not depended on by anyone -> it IS main
	// Wait: tool-b depends on tool-a, tool-c depends on tool-b
	// GetMainTools returns tools that NO OTHER tool depends on -> only tool-c
	main := repo.GetMainTools()

	if len(main) != 1 {
		t.Errorf("GetMainTools() returned %d tools, want 1", len(main))
	}
	if len(main) > 0 && main[0].Name != "tool-c" {
		t.Errorf("GetMainTools()[0].Name = %q, want %q", main[0].Name, "tool-c")
	}
}

func TestGetDependencies(t *testing.T) {
	repo := newTestRepo(t)

	tests := []struct {
		name      string
		toolName  string
		wantDeps  []string
	}{
		{"tool with no deps", "tool-a", []string{}},
		{"tool with one dep", "tool-b", []string{"tool-a"}},
		{"tool with transitive dep", "tool-c", []string{"tool-b"}},
		{"nonexistent tool", "ghost", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.GetDependencies(tt.toolName)
			if tt.wantDeps == nil {
				if got != nil {
					t.Errorf("GetDependencies(%q) = %v, want nil", tt.toolName, got)
				}
				return
			}
			if len(got) != len(tt.wantDeps) {
				t.Errorf("GetDependencies(%q) returned %d deps, want %d", tt.toolName, len(got), len(tt.wantDeps))
				return
			}
			for i, dep := range got {
				if dep.Name != tt.wantDeps[i] {
					t.Errorf("GetDependencies(%q)[%d].Name = %q, want %q", tt.toolName, i, dep.Name, tt.wantDeps[i])
				}
			}
		})
	}
}

func TestGetAllDependencies(t *testing.T) {
	repo := newTestRepo(t)

	tests := []struct {
		name     string
		toolName string
		wantLen  int
		wantDeps []string
	}{
		{"no deps", "tool-a", 0, []string{}},
		{"one dep", "tool-b", 1, []string{"tool-a"}},
		{"transitive: tool-c depends on b and a", "tool-c", 2, []string{"tool-b", "tool-a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.GetAllDependencies(tt.toolName)
			if len(got) != tt.wantLen {
				t.Errorf("GetAllDependencies(%q) returned %d deps, want %d", tt.toolName, len(got), tt.wantLen)
			}
		})
	}
}

func TestSave(t *testing.T) {
	repo := newTestRepo(t)

	t.Run("update existing tool", func(t *testing.T) {
		tool := repo.GetByID("tool-a")
		if tool == nil {
			t.Fatal("tool-a not found")
		}
		updated := *tool
		updated.Description = "Updated Description"
		repo.Save(updated)

		got := repo.GetByID("tool-a")
		if got == nil {
			t.Fatal("tool-a not found after save")
		}
		if got.Description != "Updated Description" {
			t.Errorf("Save() description = %q, want %q", got.Description, "Updated Description")
		}
		if len(repo.GetAll()) != 3 {
			t.Errorf("Save() changed tool count, want 3")
		}
	})

	t.Run("append new tool", func(t *testing.T) {
		newTool := entities.Tool{Name: "tool-new", Category: entities.CategoryTools}
		repo.Save(newTool)

		got := repo.GetByID("tool-new")
		if got == nil {
			t.Error("new tool not found after save")
		}
		if len(repo.GetAll()) != 4 {
			t.Errorf("after append, tool count = %d, want 4", len(repo.GetAll()))
		}
	})
}

func TestGetThemes(t *testing.T) {
	repo := newTestRepo(t)

	themes := repo.GetThemes()
	if len(themes) != 1 {
		t.Errorf("GetThemes() returned %d themes, want 1", len(themes))
	}
	if themes[0].Name != "test-theme" {
		t.Errorf("theme name = %q, want %q", themes[0].Name, "test-theme")
	}
}

func TestGetThemeByName(t *testing.T) {
	repo := newTestRepo(t)

	t.Run("existing theme", func(t *testing.T) {
		got := repo.GetThemeByName("test-theme")
		if got == nil {
			t.Error("GetThemeByName() = nil, want theme")
		}
	})

	t.Run("nonexistent theme", func(t *testing.T) {
		got := repo.GetThemeByName("nord")
		if got != nil {
			t.Errorf("GetThemeByName(nord) = %v, want nil", got)
		}
	})
}
