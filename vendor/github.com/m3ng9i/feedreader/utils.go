package feedreader

import "encoding/xml"
import "bytes"
import "io"
import "html"
import "strings"
import myhtml "github.com/m3ng9i/go-utils/html"


// convert content to html
func transformContent(content string) string {
    if strings.ContainsAny(content, "<>") {
        return content
    } else {
        c := html.UnescapeString(content)
        if strings.ContainsAny(c, "<>") {
            return c
        } else {
            return myhtml.Text2Html(c)
        }
    }
}


// unmarshal xml bytes to interface value v, use this function to instead of xml.Unmarshal()
func unmarshal(b []byte, v interface{}) error {
    decorder := xml.NewDecoder(bytes.NewReader(b))
    decorder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
        return input, nil
    }
    return decorder.Decode(v)
}
