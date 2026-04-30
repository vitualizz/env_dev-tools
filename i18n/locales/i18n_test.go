package locales_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/vitualizz/vitualizz-devstack/i18n/locales"
)

func TestI18n(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "I18n Suite")
}

var _ = Describe("I18nSimple", func() {
	var i18n *locales.I18nSimple

	BeforeEach(func() {
		i18n = locales.NewI18nSimple()
	})

	Describe("default language", func() {
		It("should be Spanish", func() {
			Expect(i18n.T("welcome")).To(Equal("Bienvenido a EnvSetup"))
		})
	})

	Describe("switching language", func() {
		Context("to English", func() {
			BeforeEach(func() {
				i18n.SetLanguage("en")
			})

			It("should return English translations", func() {
				Expect(i18n.T("welcome")).To(Equal("Welcome to EnvSetup"))
			})
		})

		Context("to Spanish", func() {
			BeforeEach(func() {
				i18n.SetLanguage("es")
			})

			It("should return Spanish translations", func() {
				Expect(i18n.T("welcome")).To(Equal("Bienvenido a EnvSetup"))
			})
		})

		Context("to unsupported language", func() {
			It("should fall back to Spanish", func() {
				i18n.SetLanguage("fr")
				Expect(i18n.T("welcome")).To(Equal("Bienvenido a EnvSetup"))
			})
		})
	})

	Describe("translations", func() {
		It("should translate install", func() {
			Expect(i18n.T("install")).To(Equal("Instalar"))
		})

		It("should translate exit", func() {
			Expect(i18n.T("exit")).To(Equal("Salir"))
		})

		It("should translate settings", func() {
			Expect(i18n.T("settings")).To(Equal("Configuración"))
		})
	})

	Describe("available languages", func() {
		It("should return both languages", func() {
			langs := i18n.GetAvailableLanguages()
			Expect(langs).To(ContainElements("es", "en"))
		})
	})
})