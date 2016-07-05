package http

import "fmt"
import "net/http"
import "github.com/abbot/go-http-auth"
import "golang.org/x/net/context"


type DigestAuth struct {
    Realm                   string                  // Authentication realm
    Secret                  auth.SecretProvider     // return ha1 for authentication success, return empty string for authentication failed
    ClientCacheSize         int                     // see "go-http-auth" package for more information
    ClientCacheTolerance    int                     // see "go-http-auth" package for more information
}


type ErrMsgTitleBody struct {
    Title string
    Body string
}


const html401 = `<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="initial-scale=1,width=device-width"><title>%s</title><body>%s</body></html>`


// errorHandler return a function for writing 401 error message to http.ResponseWriter
func errorHandler(err interface{}) func(w http.ResponseWriter) {

    var msg string

    if val, ok := err.(ErrMsgTitleBody); ok {
        msg = fmt.Sprintf(html401, val.Title, val.Body)
    } else if val, ok := err.(string); ok {
        msg = val
    } else {
        // default behavior. ignore err for it's incorrect type.
        msg = fmt.Sprintf(html401, "401 Unauthorized", "<h1>401 Unauthorized</h1>")
    }

    return func(w http.ResponseWriter) {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte(msg))
    }
}


/* DigestAuthHandler provide HTTP digest authentication.
If authentication failed, call failFunc function.

Parameters:
    handler     Call this http handler if authentication success.
    failMsg     If authentication failed, write the failMsg to http.ResponseWriter.
    failFunc    if authentication failed and failFunc is not nil,
                call this function before write error message to ResponseWriter.

Type of failMsg could be ErrMsgTitleBody or string.

    1. If failMsg is type of ErrMsgTitleBody, set failMsg.Title and failMsg.Body as
    html page's title and body, and write to ResponseWriter.

    2. If failMsg is type of string, write the string as html to ResponseWriter.

    3. Otherwise, set "401 Unauthorized" as html page's title and body, and write
    the html to ResponseWriter.

    See errorHandler function for more information.

Example:

    package main

    import (
        "fmt"
        "net/http"
        "time"
        "crypto/md5"

        hh "github.com/m3ng9i/go-utils/http"
    )


    func serve(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "it's ok")
    }

    func failfunc() {
        fmt.Println("login failed")
        time.Sleep(2 * time.Second)
    }

    func main() {

        fmt.Println("server start")

        auth := hh.DigestAuth {
            Realm: "This page need authentication",
            Secret: func(user, realm string) string {
                if user == "john" {
                    // password is "hello"
                    hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", user, realm, "hello")))
                    return fmt.Sprintf("%x", hash)
                }
                return ""
            },
        }

        errmsg := hh.ErrMsgTitleBody {
            Title: "401 Unauthorized",
            Body: "<h1>This page need authentication</h1>",
        }

        http.ListenAndServe(":8000", auth.DigestAuthHandler(serve, errmsg, failfunc))
    }
*/
func (a *DigestAuth) DigestAuthHandler(handler http.HandlerFunc, failMsg interface{}, failFunc func()) http.HandlerFunc {
    authenticator := auth.NewDigestAuthenticator(a.Realm, a.Secret)
    if a.ClientCacheSize > 0 {
        authenticator.ClientCacheSize = a.ClientCacheSize
    }
    if a.ClientCacheTolerance > 0 {
        authenticator.ClientCacheTolerance = a.ClientCacheTolerance
    }

    errHandler := errorHandler(failMsg)

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := authenticator.NewContext(context.Background(), r)

        authInfo := auth.FromContext(ctx)
        authInfo.UpdateHeaders(w.Header())
        if authInfo == nil || !authInfo.Authenticated {
            if failFunc != nil {
                failFunc()
            }
            errHandler(w)
            return
        }
        handler(w, r)
    })
}


// DigestAuthHandler wrap a http handler function with digest authentication.
func (a *DigestAuth) DigestAuthWrap(handler http.HandlerFunc) http.HandlerFunc {
    authenticator := auth.NewDigestAuthenticator(a.Realm, a.Secret)
    if a.ClientCacheSize > 0 {
        authenticator.ClientCacheSize = a.ClientCacheSize
    }
    if a.ClientCacheTolerance > 0 {
        authenticator.ClientCacheTolerance = a.ClientCacheTolerance
    }
    return authenticator.JustCheck(handler)
}

