[![Build Status](https://travis-ci.org/st3v/translator.svg?branch=master)](https://travis-ci.org/st3v/translator)
[![GoDoc](https://godoc.org/github.com/st3v/translator?status.png)](https://godoc.org/github.com/st3v/translator)

Translator
==========

Go package for easy access to 
[Microsoft Translator API](http://msdn.microsoft.com/en-us/library/ff512423.aspx) and 
[Google Translate API](https://cloud.google.com/translate/docs).

## Installation

```
go get github.com/st3v/translator
```

## Instantiation

### Microsoft Translator API

Sign-up for Microsoft Translator API ([see instructions](http://blogs.msdn.com/b/translation/p/gettingstarted1.aspx)) and get your developer credentials. 
Use the obtained client ID and secret to instantiate a translator as shown
below.

```go
package main

import (
  "fmt"
  "log"

  "github.com/st3v/translator/microsoft"
)

func main() {
  translator := microsoft.NewTranslator("YOUR-CLIENT-ID", "YOUR-CLIENT-SECRET")
    
  translation, err := translator.Translate("Hello World!", "en", "de")
  if err != nil {
    log.Panicf("Error during translation: %s", err.Error())
  }

  fmt.Println(translation)
}
```

### Google Translate API

Sign-up for Google Developers Console and enable the Translate API ([see instructions] (https://cloud.google.com/translate/v2/getting_started#setup)).
Obtain the API key for your application and use it to instantiate a translator
as show below.

```go
package main

import (
  "fmt"
  "log"

  "github.com/st3v/translator/google"
)

func main() {
  translator := google.NewTranslator("YOUR-GOOGLE-API-KEY")

  translation, err := translator.Translate("Hello World!", "en", "de")
  if err != nil {
    log.Panicf("Error during translation: %s", err.Error())
  }

  fmt.Println(translation)
}
```

## Translation

Use the `Translate` function to translate text from one language to another. The
function expects the caller to use API-specific language codes to specify the source
and target language for the translation. 

See [Microsoft's](https://msdn.microsoft.com/en-us/library/hh456380.aspx) or 
[Google's](https://cloud.google.com/translate/v2/using_rest#language-params)
documentation for a list of supported languages and their corresponding codes.
Or use the [`Languages`](#languages) function to programmatically obtain the list of 
supported languages and their codes.

**Signature**

```go
// Translate takes a string in a given language and returns its translation
// to another language. Source and destination languages are specified by their
// corresponding language codes.
Translate(text, from, to string) (string, error)
```

**Usage**

```go
translation, err := translator.Translate("Hello World!", "en", "de")
if err != nil {
  log.Panicf("Error during translation: %s", err.Error())
}

fmt.Printf("Translation: %s\n", translation)
```

## Language Detection

You can use the `Detect` function to detect the language of a give word or sentence.

**Signature**

```go
// Detect identifies the language of the given text and returns the
// corresponding language code.
Detect(text string) (string, error)
```

**Usage**

```go
languageCode, err := translator.Detect("¿cómo está?")
if err != nil {
  log.Panicf("Error detecting language: %s", err.Error())
}

fmt.Printf("Detected language code: %s", languageCode)
```

## Supported Languages
<a name="languages"></a>

The `Languages` function returns a list of all languages supported 
by the API you are using. The function will provide you with the english 
name and API-specific code for each language.

**Signature**

```go
// Languages returns a slice of language structs that are supported
// by the given translator.
Languages() ([]Language, error)
```

**Usage**

```go
languages, err := translator.Languages()
if err != nil {
  log.Panicf("Error getting supported languages: %s", err.Error())
}

for _, language := range languages {
  fmt.Printf("%s (%s)\n", language.Name, language.Code)
}
```

## Licensing
Translator is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/st3v/translator/blob/master/LICENSE) for the full
license text.

