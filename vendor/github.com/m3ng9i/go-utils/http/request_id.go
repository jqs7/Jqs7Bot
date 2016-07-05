package http

import "fmt"
import "time"
import "crypto/rand"
import "crypto/md5"
import "io"


// A string for identify each http request
type RequestId string


/*
Create a function to generate random request ids.
You can use parameter length to set the length of the result.
The max length of result is 32.
You can use request url as the parameter in the returned function to provide a more randomly result.

Example:
    NewReqeustId := RequestIdGenerator(12)
    fmt.Println(NewReqeustId("http://example.com"))
    // Output: b08ea4a86fe3
*/
func RequestIdGenerator(length int) func(url ...string) RequestId {

    if length < 0 {
        length = 0
    } else if length > 32 {
        length = 32
    }

    return func(url ...string) RequestId {
        if length == 0 {
            return ""
        }

        // 32 random bytes
        b := make([]byte, 32)
        rand.Read(b)

        hash := md5.New()

        var s string
        if len(url) > 0 {
            s = url[0]
        }

        io.WriteString(hash, fmt.Sprintf("%d%s", time.Now().UnixNano(), s))
        hash.Write(b)

        return RequestId(fmt.Sprintf("%x", hash.Sum(nil))[:length])
    }
}
