package installers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/vitualizz/envsetup/internal/domain/entities"
	"github.com/vitualizz/envsetup/internal/infrastructure/installers"
)

func TestInstaller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Installer Suite")
}

var _ = Describe("ToolInstaller", func() {
	var installer *installers.ToolInstaller

	BeforeEach(func() {
		installer = installers.NewToolInstaller()
	})

	Describe("creating installer", func() {
		It("should not be nil", func() {
			Expect(installer).ToNot(BeNil())
		})
	})

	Describe("checking tool installation", func() {
		Context("when tool has no check command", func() {
			It("should return false", func() {
				tool := &entities.Tool{
					Name:  "test",
					Check: "",
				}
				installed, err := installer.IsInstalled(tool)
				Expect(err).ToNot(HaveOccurred())
				Expect(installed).To(BeFalse())
			})
		})

		Context("when check command fails", func() {
			It("should return false", func() {
				tool := &entities.Tool{
					Name:  "nonexistent-tool-12345",
					Check: "echo 'not found' && exit 1",
				}
				installed, err := installer.IsInstalled(tool)
				Expect(err).ToNot(HaveOccurred())
				Expect(installed).To(BeFalse())
			})
		})

		Context("when check command succeeds", func() {
			It("should return true", func() {
				tool := &entities.Tool{
					Name:  "echo",
					Check: "echo 'hello'",
				}
				installed, err := installer.IsInstalled(tool)
				Expect(err).ToNot(HaveOccurred())
				Expect(installed).To(BeTrue())
			})
		})
	})

	Describe("installing tool", func() {
		Context("when tool has no install command for current distro", func() {
			It("should return false result", func() {
				tool := &entities.Tool{
					Name:    "test",
					Install: map[entities.Distro]string{},
				}
				result, err := installer.Install(tool)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Success).To(BeFalse())
			})
		})

		Context("when install command is invalid", func() {
			It("should return error", func() {
				tool := &entities.Tool{
					Name:        "invalid-tool",
					Install:    map[entities.Distro]string{entities.DistroAll: "exit 1"},
					Description: "Test tool",
				}
				result, err := installer.Install(tool)
				Expect(err).To(HaveOccurred())
				Expect(result).ToNot(BeNil())
			})
		})
	})
})