package entities_test

import (
	"testing"

	"github.com/vitualizz/envsetup/internal/domain/entities"
)

func TestGetInstallCmd(t *testing.T) {
	arch := "pacman -S foo"
	deb := "apt-get install -y foo"
	all := "curl install.sh | sh"
	fallback := "cargo install foo"

	tests := []struct {
		name    string
		install map[entities.Distro]string
		distro  entities.Distro
		want    string
	}{
		{
			name:    "exact distro match",
			install: map[entities.Distro]string{entities.DistroArch: arch},
			distro:  entities.DistroArch,
			want:    arch,
		},
		{
			name:    "all key used when no exact match",
			install: map[entities.Distro]string{entities.DistroAll: all},
			distro:  entities.DistroAlpine,
			want:    all,
		},
		{
			name:    "exact match wins over all",
			install: map[entities.Distro]string{entities.DistroArch: arch, entities.DistroAll: all},
			distro:  entities.DistroArch,
			want:    arch,
		},
		{
			name:    "detection order fallback to debian",
			install: map[entities.Distro]string{entities.DistroDebian: deb},
			distro:  entities.DistroAlpine,
			want:    deb,
		},
		{
			name:    "fallback key used as last resort",
			install: map[entities.Distro]string{entities.DistroFallback: fallback},
			distro:  entities.DistroAlpine,
			want:    fallback,
		},
		{
			name:    "all wins over detection order",
			install: map[entities.Distro]string{entities.DistroAll: all, entities.DistroDebian: deb},
			distro:  entities.DistroAlpine,
			want:    all,
		},
		{
			name:    "empty install map returns empty string",
			install: map[entities.Distro]string{},
			distro:  entities.DistroAlpine,
			want:    "",
		},
		{
			name:    "nil install map returns empty string",
			install: nil,
			distro:  entities.DistroArch,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &entities.Tool{Name: "test", Install: tt.install}
			got := tool.GetInstallCmd(tt.distro)
			if got != tt.want {
				t.Errorf("GetInstallCmd() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetUninstallCmd(t *testing.T) {
	arch := "pacman -R foo"
	all := "rm -rf ~/.local/bin/foo"

	tests := []struct {
		name      string
		uninstall map[entities.Distro]string
		distro    entities.Distro
		want      string
	}{
		{
			name:      "exact match",
			uninstall: map[entities.Distro]string{entities.DistroArch: arch},
			distro:    entities.DistroArch,
			want:      arch,
		},
		{
			name:      "all fallback",
			uninstall: map[entities.Distro]string{entities.DistroAll: all},
			distro:    entities.DistroFedora,
			want:      all,
		},
		{
			name:      "empty returns empty",
			uninstall: map[entities.Distro]string{},
			distro:    entities.DistroArch,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &entities.Tool{Name: "test", Uninstall: tt.uninstall}
			got := tool.GetUninstallCmd(tt.distro)
			if got != tt.want {
				t.Errorf("GetUninstallCmd() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasInstallCommand(t *testing.T) {
	tests := []struct {
		name    string
		install map[entities.Distro]string
		want    bool
	}{
		{
			name:    "has command",
			install: map[entities.Distro]string{entities.DistroAll: "echo install"},
			want:    true,
		},
		{
			name:    "empty map",
			install: map[entities.Distro]string{},
			want:    false,
		},
		{
			name:    "nil map",
			install: nil,
			want:    false,
		},
		{
			name:    "only empty string value",
			install: map[entities.Distro]string{entities.DistroAll: ""},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &entities.Tool{Name: "test", Install: tt.install}
			if got := tool.HasInstallCommand(); got != tt.want {
				t.Errorf("HasInstallCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBundle(t *testing.T) {
	tests := []struct {
		name      string
		install   map[entities.Distro]string
		dependsOn []string
		want      bool
	}{
		{
			name:    "tool with install command is not a bundle",
			install: map[entities.Distro]string{entities.DistroAll: "echo install"},
			want:    false,
		},
		{
			name:      "tool with dependencies is not a bundle",
			install:   map[entities.Distro]string{},
			dependsOn: []string{"other-tool"},
			want:      false,
		},
		{
			name:    "tool with no install and no deps is a bundle",
			install: map[entities.Distro]string{},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &entities.Tool{
				Name:      "test",
				Install:   tt.install,
				DependsOn: tt.dependsOn,
			}
			if got := tool.IsBundle(); got != tt.want {
				t.Errorf("IsBundle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategoryIsValid(t *testing.T) {
	tests := []struct {
		cat  entities.Category
		want bool
	}{
		{entities.CategoryTerminal, true},
		{entities.CategoryShell, true},
		{entities.CategoryEditor, true},
		{entities.CategoryTools, true},
		{entities.CategoryContainer, true},
		{entities.CategoryFonts, true},
		{entities.Category("theme"), false},
		{entities.Category("unknown"), false},
		{entities.Category(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.cat), func(t *testing.T) {
			if got := tt.cat.IsValid(); got != tt.want {
				t.Errorf("Category(%q).IsValid() = %v, want %v", tt.cat, got, tt.want)
			}
		})
	}
}

func TestDistroDetectionOrder(t *testing.T) {
	// Verifica que fallback esté al final del detection order
	order := entities.DistroDetectionOrder
	if len(order) == 0 {
		t.Fatal("DistroDetectionOrder is empty")
	}
	last := order[len(order)-1]
	if last != entities.DistroFallback {
		t.Errorf("last element of DistroDetectionOrder = %q, want %q", last, entities.DistroFallback)
	}
}
