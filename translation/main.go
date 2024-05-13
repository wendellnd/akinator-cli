package translation

import "fmt"

type languages []string

func (l languages) IsValid(language string) bool {
	for _, lang := range l {
		if lang == language {
			return true
		}
	}

	return false
}

func (l languages) GetDefault() string {
	return l[0]
}

var SupportedLanguages = languages{
	"en",
	"pt",
}

var SupportedLanguagesMap = map[string]map[string]string{
	"pt": {
		"Yes":                  "Sim",
		"No":                   "Não",
		"Starting Akinator...": "Inicializando Akinator...",
		"Akinator thinks it's": "Akinator acha que é",
		"Thanks for playing!":  "Agradeçemos por jogar!",
	},
}

type Translation struct {
	language string
}

func NewTranslation(language string) (translation *Translation) {
	isValid := SupportedLanguages.IsValid(language)
	if !isValid {
		// Only alerts if the flag is set
		if language != "" {
			fmt.Println("Invalid language, using default language")
		}
		language = SupportedLanguages.GetDefault()
	}

	return &Translation{
		language: language,
	}
}

func (t *Translation) SetLanguage(language string) {
	isValid := SupportedLanguages.IsValid(language)
	if !isValid {
		language = SupportedLanguages.GetDefault()
	}

	t.language = language
}

func (t *Translation) Translate(key string) string {
	if t.language == SupportedLanguages.GetDefault() {
		return key
	}

	return SupportedLanguagesMap[t.language][key]
}
