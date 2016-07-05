package microsoft

import (
	"github.com/st3v/translator"
	msauth "github.com/st3v/translator/microsoft/auth"
)

type api struct {
	languageCatalog     LanguageCatalog
	translationProvider TranslationProvider
}

// NewTranslator returns a struct that implements the Translator
// interface by exposing a Translate and a Languages function that
// are backed by Microsoft's translation API.
// The function takes the clientID and clientSecret for an existing
// app registered in Microsoft's Azure DataMarket.
func NewTranslator(clientID, clientSecret string) translator.Translator {
	scope := "http://api.microsofttranslator.com"
	router := newRouter()
	authenticator := msauth.NewAuthenticator(clientID, clientSecret, scope, router.AuthURL())
	return &api{
		languageCatalog:     newLanguageCatalog(newLanguageProvider(authenticator, router)),
		translationProvider: newTranslationProvider(authenticator, router),
	}
}

func (a *api) Translate(text, from, to string) (string, error) {
	return a.translationProvider.Translate(text, from, to)
}

func (a *api) Languages() ([]translator.Language, error) {
	return a.languageCatalog.Languages()
}

func (a *api) Detect(text string) (string, error) {
	return a.translationProvider.Detect(text)
}
