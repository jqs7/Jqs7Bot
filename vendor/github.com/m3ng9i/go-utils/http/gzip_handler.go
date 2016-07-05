package http

import "net/http"
import "strings"
import "io"
import "compress/gzip"
import "path"
import "sync"
import "github.com/m3ng9i/go-utils/possible"


var gzipWriterPool = sync.Pool {
    New: func() interface{} {
        return gzip.NewWriter(nil)
    },
}


type GzipResponseWriter struct {
    io.Writer
    http.ResponseWriter
}


func (w GzipResponseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}


/* Wrap a handler to a gzip handler.

The handler first check the Accept-Encoding request header to verify if the client supports gzip compression,
if not, the handler send the orignial (not compressed) data to the client.

If checkQuery is true, the function do the following check:
    1. If the query string contains gzip=true, the handler send compressed data to the client.
    2. If the query string contains gzip=false, the handler send the orignial data to the client.
    3. If the query string does not contains gzip or gzip's value is neither true nor false,
       the handler follow the next step.

If checkName is true, the handler check the request path's file extension and determine if the path could be
compressed. If false, send orignial data to client, otherwise send the compressed data to the client.
Example: .html, .css, .txt files could be compressed but .jpg, .png, .avi does not need to be compressed.

Parameter checkQuery has higher priority than checkName.
*/
func GzipHandler(fn http.HandlerFunc, checkQuery, checkName bool) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip") {
            fn(w, r)
            return
        }

        if checkQuery {
            query := strings.ToLower(r.URL.Query().Get("gzip"))
            if query == "true" {
                goto DoGzip
            } else if query == "false" {
                fn(w, r)
                return
            }
        }

        if checkName && CanBeCompressed(path.Base(r.URL.Path)) == possible.No {
            fn(w, r)
            return
        }

        DoGzip:

        w.Header().Set("Content-Encoding", "gzip")

        gz := gzipWriterPool.Get().(*gzip.Writer)
        gz.Reset(w)
        defer func() {
            gz.Close()
            gzipWriterPool.Put(gz)
        }()

        gzWriter := GzipResponseWriter{Writer: gz, ResponseWriter: w}
        fn(gzWriter, r)
    })
}
