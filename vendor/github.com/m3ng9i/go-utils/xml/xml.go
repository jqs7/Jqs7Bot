package xml

import "regexp"

/*
Remove invalid xml characters.

When parsing some xml doc, encoding/xml package will return an error:
    XML syntax error on line XXX: illegal character code U+XXXX
This function will solve the problem.

The list of valid characters is in the XML specification:
http://www.w3.org/TR/xml/#charsets

Valid characters:
#x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | [#x10000-#x10FFFF]
*/
func RemoveInvalidChars(b []byte) []byte {
    re := regexp.MustCompile("[^\x09\x0A\x0D\x20-\uD7FF\uE000-\uFFFD\u10000-\u10FFFF]")
    return re.ReplaceAll(b, []byte{})
}



