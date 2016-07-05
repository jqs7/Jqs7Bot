package microsoft

import (
	"github.com/st3v/tracerr"
	"github.com/st3v/translator"
)

// The LanguageCatalog provides a slice of languages representing all
// languages supported by Microsoft's Translation API.
type LanguageCatalog interface {
	Languages() ([]translator.Language, error)
}

type languageCatalog struct {
	provider  LanguageProvider
	languages []translator.Language
}

func newLanguageCatalog(provider LanguageProvider) LanguageCatalog {
	return &languageCatalog{
		provider: provider,
	}
}

func (c *languageCatalog) Languages() ([]translator.Language, error) {
	if c.languages == nil {
		codes, err := c.provider.Codes()
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		names, err := c.provider.Names(codes)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		for i := range codes {
			c.languages = append(
				c.languages,
				translator.Language{
					Code: codes[i],
					Name: names[i],
				})
		}
	}
	return c.languages, nil
}
