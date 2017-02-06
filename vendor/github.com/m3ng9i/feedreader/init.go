package feedreader

import httphelper "github.com/m3ng9i/go-utils/http"


var defaultFetcher *httphelper.Fetcher


func init() {
    header := make(map[string]string)
    header["User-Agent"] = "feedreader (http://github.com/m3ng9i/feedreader)"

    defaultFetcher = httphelper.NewFetcher(nil, header)
}

