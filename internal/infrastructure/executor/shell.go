package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/vitualizz/envsetup/internal/domain/entities"
)

type Distro string

const (
	DistroUbuntu   Distro = "ubuntu"
	DistroDebian   Distro = "debian"
	DistroArch     Distro = "arch"
	DistroFedora   Distro = "fedora"
	DistroAlpine   Distro = "alpine"
	DistroUnknown  Distro = "unknown"
)

type ShellExecutor struct{}

func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

func (e *ShellExecutor) DetectDistro() Distro {
	if _, err := os.Stat("/etc/os-release"); err != nil {
		return DistroUnknown
	}

	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return DistroUnknown
	}

	content := string(data)
	if strings.Contains(content, "ID=ubuntu") {
		return DistroUbuntu
	}
	if strings.Contains(content, "ID=debian") {
		return DistroDebian
	}
	if strings.Contains(content, "ID=arch") || strings.Contains(content, "ID=archlinux") {
		return DistroArch
	}
	if strings.Contains(content, "ID=fedora") {
		return DistroFedora
	}
	if strings.Contains(content, "ID=alpine") {
		return DistroAlpine
	}

	return DistroUnknown
}

func (e *ShellExecutor) GetPkgManager() string {
	distro := e.DetectDistro()

	switch distro {
	case DistroUbuntu, DistroDebian:
		return "apt-get"
	case DistroArch:
		return "pacman"
	case DistroFedora:
		return "dnf"
	case DistroAlpine:
		return "apk"
	}

	if _, err := exec.LookPath("apt-get"); err == nil {
		return "apt-get"
	}
	if _, err := exec.LookPath("pacman"); err == nil {
		return "pacman"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "dnf"
	}
	if _, err := exec.LookPath("apk"); err == nil {
		return "apk"
	}

	return "unknown"
}

func (e *ShellExecutor) Execute(cmd string) (string, error) {
	start := time.Now()

	sh := "/bin/sh"
	c := "-c"
	if isZshAvailable() {
		sh = "/bin/zsh"
	}

	command := exec.Command(sh, c, cmd)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return "", fmt.Errorf("command failed: %w - %s", err, stderr.String())
	}

	return fmt.Sprintf("[%dms] %s", duration, stdout.String()), nil
}

func (e *ShellExecutor) ExecuteWithOutput(cmd string) (string, error) {
	sh := "/bin/sh"
	c := "-c"
	if isZshAvailable() {
		sh = "/bin/zsh"
	}

	command := exec.Command(sh, c, cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

func isZshAvailable() bool {
	_, err := exec.LookPath("zsh")
	return err == nil
}

var _ entities.Executor = (*ShellExecutor)(nil)