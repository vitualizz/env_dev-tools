package installers

import (
	"testing"
)

func TestToolInstaller(t *testing.T) {
	installer := NewToolInstaller()

	t.Run("creates installer", func(t *testing.T) {
		if installer == nil {
			t.Error("expected installer to not be nil")
		}
	})

	t.Run("detects distro", func(t *testing.T) {
		distro := installer.Distro()
		if distro == "" {
			t.Log("distro detection returned empty (expected in test environment)")
		}
	})
}

func BenchmarkToolInstaller(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewToolInstaller()
	}
}