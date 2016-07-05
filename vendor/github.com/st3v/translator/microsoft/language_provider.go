package microsoft

import (
	"encoding/xml"
	"io/ioutil"
	"strings"

	"github.com/st3v/tracerr"
	"github.com/st3v/translator/http"
)

// The LanguageProvider retrieves the names and codes of all languages
// supported by Microsoft's Translation API.
type LanguageProvider interface {
	Codes() ([]string, error)
	Names(codes []string) ([]string, error)
}

type languageProvider struct {
	router     Router
	httpClient http.Client
}

func newLanguageProvider(authenticator http.Authenticator, router Router) LanguageProvider {
	return &languageProvider{
		router:     router,
		httpClient: http.NewClient(authenticator),
	}
}

func (p *languageProvider) Names(codes []string) ([]string, error) {
	payload, _ := xml.Marshal(newXMLArrayOfStrings(codes))
	uri := p.router.LanguageNamesURL() + "?locale=en"

	response, err := p.httpClient.SendRequest("POST", uri, strings.NewReader(string(payload)), "text/xml")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	result := &xmlArrayOfStrings{}
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, tracerr.Wrap(err)
	}

	return result.Strings, nil
}

func (p *languageProvider) Codes() ([]string, error) {
	response, err := p.httpClient.SendRequest("GET", p.router.LanguageCodesURL(), nil, "text/plain")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	result := &xmlArrayOfStrings{}
	if err = xml.Unmarshal(body, &result); err != nil {
		return nil, tracerr.Wrap(err)
	}

	return result.Strings, nil
}
