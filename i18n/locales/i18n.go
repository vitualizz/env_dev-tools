package locales

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/vitualizz/vitualizz-devstack/internal/domain/interfaces"
)

type Translation struct {
	Welcome          string
	SelectTools      string
	Install          string
	Uninstall        string
	InstallAll       string
	UninstallAll     string
	Exit             string
	Back             string
	Settings         string
	Language         string
	Categories       string
	Terminal         string
	Shell            string
	Tools            string
	Container        string
	Languages        string
	Utils            string
	Search           string
	NoToolsFound     string
	Installing       string
	Uninstalling     string
	Success          string
	Failed           string
	AlreadyInstalled string
	NotInstalled     string
	PressEnter       string
	Confirm          string
	Cancel           string
	Progress         string
	Completed        string
	Skipped          string
	Error            string
	Required         string
	Optional         string
	SelectCategory   string
	SelectAction     string
	Results          string
	Summary          string
	Total            string
	Installed        string
	FailedCount      string
}

type I18n struct {
	translations map[string]Translation
	currentLang  string
	defaultLang  string
}

func NewI18n(localesPath string) (*I18n, error) {
	i18n := &I18n{
		translations: make(map[string]Translation),
		defaultLang:  "es",
		currentLang:  "es",
	}

	data, err := os.ReadFile(filepath.Join(localesPath, "es.json"))
	if err == nil {
		i18n.translations["es"] = parseTranslation(data)
	}

	data, err = os.ReadFile(filepath.Join(localesPath, "en.json"))
	if err == nil {
		i18n.translations["en"] = parseTranslation(data)
	}

	return i18n, nil
}

func parseTranslation(data []byte) Translation {
	trans := Translation{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "\"") || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.Trim(parts[0], "\" ")
		value := strings.Trim(parts[1], ",\" \n")

		switch key {
		case "welcome": trans.Welcome = value
		case "select_tools": trans.SelectTools = value
		case "install": trans.Install = value
		case "uninstall": trans.Uninstall = value
		case "install_all": trans.InstallAll = value
		case "uninstall_all": trans.UninstallAll = value
		case "exit": trans.Exit = value
		case "back": trans.Back = value
		case "settings": trans.Settings = value
		case "language": trans.Language = value
		case "categories": trans.Categories = value
		case "terminal": trans.Terminal = value
		case "shell": trans.Shell = value
		case "tools": trans.Tools = value
		case "container": trans.Container = value
		case "languages": trans.Languages = value
		case "utils": trans.Utils = value
		case "search": trans.Search = value
		case "no_tools_found": trans.NoToolsFound = value
		case "installing": trans.Installing = value
		case "uninstalling": trans.Uninstalling = value
		case "success": trans.Success = value
		case "failed": trans.Failed = value
		case "already_installed": trans.AlreadyInstalled = value
		case "not_installed": trans.NotInstalled = value
		case "press_enter": trans.PressEnter = value
		case "confirm": trans.Confirm = value
		case "cancel": trans.Cancel = value
		case "progress": trans.Progress = value
		case "completed": trans.Completed = value
		case "skipped": trans.Skipped = value
		case "error": trans.Error = value
		case "required": trans.Required = value
		case "optional": trans.Optional = value
		case "select_category": trans.SelectCategory = value
		case "select_action": trans.SelectAction = value
		case "results": trans.Results = value
		case "summary": trans.Summary = value
		case "total": trans.Total = value
		case "installed": trans.Installed = value
		case "failed_count": trans.FailedCount = value
		}
	}
	return trans
}

func (i *I18n) Get(key string, lang string) string {
	trans, ok := i.translations[lang]
	if !ok {
		trans, ok = i.translations[i.defaultLang]
		if !ok {
			return key
		}
	}

	switch key {
	case "welcome": return trans.Welcome
	case "select_tools": return trans.SelectTools
	case "install": return trans.Install
	case "uninstall": return trans.Uninstall
	case "install_all": return trans.InstallAll
	case "uninstall_all": return trans.UninstallAll
	case "exit": return trans.Exit
	case "back": return trans.Back
	case "settings": return trans.Settings
	case "language": return trans.Language
	case "categories": return trans.Categories
	case "terminal": return trans.Terminal
	case "shell": return trans.Shell
	case "tools": return trans.Tools
	case "container": return trans.Container
	case "languages": return trans.Languages
	case "utils": return trans.Utils
	case "search": return trans.Search
	case "no_tools_found": return trans.NoToolsFound
	case "installing": return trans.Installing
	case "uninstalling": return trans.Uninstalling
	case "success": return trans.Success
	case "failed": return trans.Failed
	case "already_installed": return trans.AlreadyInstalled
	case "not_installed": return trans.NotInstalled
	case "press_enter": return trans.PressEnter
	case "confirm": return trans.Confirm
	case "cancel": return trans.Cancel
	case "progress": return trans.Progress
	case "completed": return trans.Completed
	case "skipped": return trans.Skipped
	case "error": return trans.Error
	case "required": return trans.Required
	case "optional": return trans.Optional
	case "select_category": return trans.SelectCategory
	case "select_action": return trans.SelectAction
	case "results": return trans.Results
	case "summary": return trans.Summary
	case "total": return trans.Total
	case "installed": return trans.Installed
	case "failed_count": return trans.FailedCount
	}
	return key
}

func (i *I18n) SetLanguage(lang string) {
	if _, ok := i.translations[lang]; ok {
		i.currentLang = lang
	}
}

func (i *I18n) GetCurrentLanguage() string {
	return i.currentLang
}

func (i *I18n) GetAvailableLanguages() []string {
	langs := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		langs = append(langs, lang)
	}
	return langs
}

func (i *I18n) T(key string) string {
	return i.Get(key, i.currentLang)
}

var _ interfaces.I18nPort = (*I18n)(nil)

type I18nSimple struct {
	lang string
	translations map[string]map[string]string
}

func NewI18nSimple() *I18nSimple {
	return &I18nSimple{
		lang: "es",
		translations: map[string]map[string]string{
			"es": {
				"welcome":           "Bienvenido a EnvSetup",
				"select_tools":      "Selecciona las herramientas a instalar",
				"install":           "Instalar",
				"uninstall":         "Desinstalar",
				"install_all":       "Instalar todo",
				"uninstall_all":     "Desinstalar todo",
				"exit":              "Salir",
				"back":              "Volver",
				"settings":          "Configuración",
				"language":          "Idioma",
				"categories":        "Categorías",
				"terminal":          "Terminal",
				"shell":             "Shell",
				"tools":             "Herramientas",
				"container":         "Contenedores",
				"languages":         "Lenguajes",
				"utils":             "Utilidades",
				"search":            "Buscar",
				"no_tools_found":    "No se encontraron herramientas",
				"installing":        "Instalando",
				"uninstalling":      "Desinstalando",
				"success":           "Éxito",
				"failed":            "Fallido",
				"already_installed": "Ya instalado",
				"not_installed":     "No instalado",
				"not_installed_yet": "No instalado",
				"press_enter":       "Presiona Enter para continuar",
				"confirm":           "Confirmar",
				"cancel":            "Cancelar",
				"progress":          "Progreso",
				"completed":         "Completado",
				"skipped":           "Omitido",
				"error":             "Error",
				"required":          "Requerido",
				"optional":          "Opcional",
				"select_category":   "Seleccionar categoría",
				"select_action":     "Seleccionar acción",
				"results":           "Resultados",
				"summary":           "Resumen",
				"total":             "Total",
				"installed":         "Instalados",
				"failed_count":      "Fallidos",
				"select_option":     "Selecciona una opción",
			},
			"en": {
				"welcome":           "Welcome to EnvSetup",
				"select_tools":      "Select tools to install",
				"install":           "Install",
				"uninstall":         "Uninstall",
				"install_all":       "Install all",
				"uninstall_all":     "Uninstall all",
				"exit":              "Exit",
				"back":              "Back",
				"settings":          "Settings",
				"language":          "Language",
				"categories":        "Categories",
				"terminal":          "Terminal",
				"shell":             "Shell",
				"tools":             "Tools",
				"container":         "Containers",
				"languages":         "Languages",
				"utils":             "Utilities",
				"search":            "Search",
				"no_tools_found":    "No tools found",
				"installing":        "Installing",
				"uninstalling":      "Uninstalling",
				"success":           "Success",
				"failed":            "Failed",
				"already_installed": "Already installed",
				"not_installed":     "Not installed",
				"not_installed_yet": "Not installed yet",
				"press_enter":       "Press Enter to continue",
				"confirm":           "Confirm",
				"cancel":            "Cancel",
				"progress":          "Progress",
				"completed":         "Completed",
				"skipped":           "Skipped",
				"error":             "Error",
				"required":          "Required",
				"optional":          "Optional",
				"select_category":   "Select category",
				"select_action":     "Select action",
				"results":           "Results",
				"summary":           "Summary",
				"total":             "Total",
				"installed":         "Installed",
				"failed_count":      "Failed",
				"select_option":     "Select an option",
			},
		},
	}
}

func (i *I18nSimple) Get(key string, lang string) string {
	if trans, ok := i.translations[lang]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}
	return key
}

func (i *I18nSimple) SetLanguage(lang string) {
	if _, ok := i.translations[lang]; ok {
		i.lang = lang
	}
}

func (i *I18nSimple) GetAvailableLanguages() []string {
	langs := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		langs = append(langs, lang)
	}
	return langs
}

func (i *I18nSimple) T(key string) string {
	return i.Get(key, i.lang)
}