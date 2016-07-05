package http

import "strings"
import "path"
import "mime"
import "github.com/m3ng9i/go-utils/possible"


var canBeCompressed map[string]bool


func init() {
    canBeCompressed = make(map[string]bool)
    canBeCompressed[".txt"]         = true      // txt
    canBeCompressed[".js"]          = true
    canBeCompressed[".json"]        = true
    canBeCompressed[".xml"]         = true
    canBeCompressed[".html"]        = true
    canBeCompressed[".htm"]         = true
    canBeCompressed[".css"]         = true
    canBeCompressed[".log"]         = true
    canBeCompressed[".md"]          = true
    canBeCompressed[".mkd"]         = true
    canBeCompressed[".markdown"]    = true
    canBeCompressed[".bmp"]         = true      // picture
    canBeCompressed[".jpg"]         = false     // picture
    canBeCompressed[".jpeg"]        = false
    canBeCompressed[".jpe"]         = false
    canBeCompressed[".png"]         = false
    canBeCompressed[".gif"]         = false
    canBeCompressed[".zip"]         = false     // zip like files
    canBeCompressed[".rar"]         = false
    canBeCompressed[".7z"]          = false
    canBeCompressed[".arj"]         = false
    canBeCompressed[".bz2"]         = false
    canBeCompressed[".bzip2"]       = false
    canBeCompressed[".gz"]          = false
    canBeCompressed[".gzip"]        = false
    canBeCompressed[".tgz"]         = false
    canBeCompressed[".z"]           = false
    canBeCompressed[".epub"]        = false     // compressed files
    canBeCompressed[".docx"]        = false
    canBeCompressed[".pptx"]        = false
    canBeCompressed[".xlsx"]        = false
    canBeCompressed[".rm"]          = false     // video
    canBeCompressed[".rmvb"]        = false
    canBeCompressed[".avi"]         = false
    canBeCompressed[".mpg"]         = false
    canBeCompressed[".mpeg"]        = false
    canBeCompressed[".mov"]         = false
    canBeCompressed[".mp4"]         = false
    canBeCompressed[".divx"]        = false
    canBeCompressed[".wmv"]         = false
    canBeCompressed[".mkv"]         = false
    canBeCompressed[".flv"]         = false
    canBeCompressed[".ra"]          = false     // audio
    canBeCompressed[".mp3"]         = false
    canBeCompressed[".wma"]         = false
    canBeCompressed[".m4a"]         = false
    canBeCompressed[".ogg"]         = false
}


// Check a filename's extension and determine if it can be compressed.
// This is used to decide whether or not to turn on gzip compression when serving static files in a web server.
func CanBeCompressed(filename string) possible.Possible {
    ext := strings.ToLower(path.Ext(filename))

    can, ok := canBeCompressed[ext]
    if ok {
        if can {
            return possible.Yes
        } else {
            return possible.No
        }
    }

    tp := strings.ToLower(mime.TypeByExtension(ext))

    for _, value := range []string { "text/", "+xml" } {
        if strings.Contains(tp, value) {
            return possible.Yes
        }
    }

    for _, value := range []string { "video/", "audio/", "image/", "compressed" } {
        if strings.Contains(tp, value) {
            return possible.No
        }
    }

    return possible.Maybe
}
