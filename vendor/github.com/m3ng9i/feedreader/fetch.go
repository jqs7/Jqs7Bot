package feedreader

import "errors"
import "regexp"
import "github.com/m3ng9i/go-utils/xml"
import httphelper "github.com/m3ng9i/go-utils/http"
import enc "github.com/m3ng9i/go-utils/encoding"


// Fetch feed content from a url, return []byte
// If returned error is not nil, it will be FetchError.
func FetchByte(url string, fetcher ...*httphelper.Fetcher) (data []byte, err error) {

    var ft *httphelper.Fetcher
    if len(fetcher) > 0 && fetcher[0] != nil {
        ft = fetcher[0]
    } else {
        ft = defaultFetcher
    }
    if ft == nil {
        err = errors.New("Fetcher is nil")
        return
    }

    var e error

    data, e = ft.FetchAll(url)
    if e != nil {
        err = &FetchError{Url: url, Err: e}
        return
    }

    // Check if the first 80 bytes of the xml doc contains 'gbk' or 'gb2312'
    // just like: <?xml version="1.0" encoding="gbk"?>, or <?xml version="1.0" encoding="gb2312"?>
    // if true, convert the bytes to utf-8.
    isGbk, _ := regexp.Match("(?i)gbk|gb2312", data[:80])
    if isGbk {
        // convert to utf-8
        data, err = enc.GbkToUtf8(data)
        if err != nil {
            return
        }
    }

    // Remove invalid xml characters
    data = xml.RemoveInvalidChars(data)

    return
}


// fetch feed content from a url, return string
// If returned error is not nil, it will be FetchError.
func FetchString(url string, fetcher ...*httphelper.Fetcher) (s string, err error) {

    var ft *httphelper.Fetcher
    if len(fetcher) > 0 && fetcher[0] != nil {
        ft = fetcher[0]
    } else {
        ft = defaultFetcher
    }
    if ft == nil {
        err = errors.New("Fetcher is nil")
        return
    }

    data, err := FetchByte(url, ft)
    if err != nil {
        return
    }

    s = string(data)

    return
}


// Grap rss or atom feed and return a *Feed struct
// If returned error is not nil, it will be FetchError or ParseError.
func Fetch(feedlink string, fetcher ...*httphelper.Fetcher) (feed *Feed, err error) {

    // If err is not nil, it will be FetchError.
    data, err := FetchByte(feedlink, fetcher...)
    if err != nil {
        return
    }

    // If err is not nil, it will be ParseError.
    feed, err = Parse(data, feedlink)
    return
}
