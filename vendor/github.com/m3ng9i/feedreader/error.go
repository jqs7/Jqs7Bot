package feedreader

import "fmt"

type FetchError struct {
    Url string
    Err error
}

type ParseError struct {
    Err error
}

func (this *FetchError) Error() string {
    return fmt.Sprintf("Error occurs when fetching feed from '%s': %s", this.Url, this.Err.Error())
}

func (this *ParseError) Error() string {
    return this.Err.Error()
}
