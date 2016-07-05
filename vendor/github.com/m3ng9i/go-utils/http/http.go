package http

import "net"
import "net/http"
import "net/url"
import "strings"
import "mime"
import "path/filepath"
import "os"
import "io"
import "fmt"


// Get client IP.
func GetIP(r *http.Request) string {
    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    return host
}


/*
Get a value of key in query string.

You should make sure to call (*http.Request).ParseForm() first, then to call this function.

Space before and after the value will be striped.
If the key appears more than once, the last value will be get.

parameters:
    r       *http.Request
    key     Key in query string
    defval  Default value of key, if not found the key, return the default value. If not provide defval, empty string will be used.
*/
func QueryValue(r *http.Request, key string, defval ...string) string {
    q := r.Form[key]
    length := len(q)
    if length == 0 {
        if len(defval) > 0 {
            return defval[0]
        } else {
            return ""
        }
    }
    return strings.TrimSpace(q[length -1])
}


// Determine a file's content type, like: text/plain; charset=utf-8
// Inspired by go's serveContent() in net/http package.
// More information: src/pkg/net/http/fs.go
func FileContentType(filename string) (string, error) {
    ctype := mime.TypeByExtension(filepath.Ext(filename))
    if ctype == "" {

        file, err := os.Open(filename)
        if err != nil {
            return "", err
        }
        defer file.Close()

        ctype, err = ContentType(file)
        if err != nil {
            return "", err
        }

    }

    return ctype, nil
}


// Determine a data stream's content type, like: text/plain; charset=utf-8
// Inspired by go's serveContent() in net/http package.
// More information: src/pkg/net/http/fs.go
func ContentType(reader io.ReadSeeker) (string, error) {
    data := make([]byte, 512)

    _, err := reader.Seek(0, os.SEEK_SET)
    if err != nil {
        return "", err
    }

    n, err := reader.Read(data)
    if err != nil {
        return "", err
    }

    return http.DetectContentType(data[:n]), nil
}


// Write download header to the response writer. filename should not contains path.
func WriteDownloadHeader(w http.ResponseWriter, filename string) {
    filename = url.QueryEscape(filename)

    // more about Content-Disposition, see rfc6226: http://tools.ietf.org/html/rfc6266
    w.Header().Set("Content-Disposition",
        fmt.Sprintf(`attachment; filename="%s"; filename*=utf-8''%s`, filename, filename))

    w.Header().Set("Content-Type", "application/octet-stream")
}
