package http

import "bytes"
import "net/http"


// ResponseSniffer is used to record http response body, status code
// and number of bytes write to the response writer.
type ResponseSniffer struct {
    Code        int                 // http status code
    Body        *bytes.Buffer       // http response body
    Size        int                 // number of bytes write to the response writer
    RecordBody  bool                // if need to record response body
    rw          http.ResponseWriter // the response writer
    wroteHeader bool
}


// Make a new response sniffer, if recordBody is true, record response body to Body, this will cost more memory.
func NewSniffer(rw http.ResponseWriter, recordBody bool) *ResponseSniffer {
    sniffer := &ResponseSniffer {
        Code        : 200,
        RecordBody  : recordBody,
        rw          : rw,
    }
    if !recordBody {
        sniffer.Body = new(bytes.Buffer)
    }
    return sniffer
}


func (this *ResponseSniffer) Header() http.Header {
    return this.rw.Header()
}


func (this *ResponseSniffer) Write(buf []byte) (int, error) {
    if !this.wroteHeader {
        this.WriteHeader(200)
        this.Code = 200
    }

    if this.RecordBody {
        if this.Body == nil {
            this.Body = new(bytes.Buffer)
        }
        this.Body.Write(buf)
    }

    size, err := this.rw.Write(buf)
    this.Size += size
    return size, err
}


func (this *ResponseSniffer) WriteHeader(code int) {
    if !this.wroteHeader {
        this.Code = code
        this.rw.WriteHeader(code)
    }
    this.wroteHeader = true
}

