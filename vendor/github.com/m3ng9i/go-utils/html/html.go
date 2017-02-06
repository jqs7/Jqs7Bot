package html

import "html"
import "strings"
import "net/url"
import "bytes"
import "github.com/PuerkitoBio/goquery"

/* Convert text to html

Example:

text := `
First line

Second line
Third line
    
<b id="html">line contains html</b>
`
html := Text2Html(text)

now html will be:
<p>First line</p><p>Second line<br>Third line</p><p>&lt;b id=&#34;html&#34;&gt;line contains html&lt;/b&gt;</p>

*/
func Text2Html(text string) string {
    text = html.EscapeString(text)
    var h []string

    newPara := true
    for _, line := range(strings.Split(text, "\n")) {
        l := strings.Trim(line, "\r\n\t ")
        if newPara {
            if l == "" {
                continue
            }
            newPara = false
            h = append(h, "<p>" + l)
        } else {
            if l == "" {
                h = append(h, "</p>")
                newPara = true
            } else {
                h = append(h, "<br>" + l)
            }
        }
    }

    if newPara == false {
        h = append(h, "</p>")
    }

    return strings.Join(h, "")
}


// Convert a relative url to an absolute url base on a base url.
// If rurl is an absolute url, the function just return it.
func AbsUrl(baseurl, rurl string) (u string, err error) {

    base, err := url.Parse(baseurl)
    if err != nil {
        return
    }

    r, err := url.Parse(rurl)
    if err != nil {
        return
    }

    u = base.ResolveReference(r).String()
    return
}


/* Convert an html doc's urls from relative to absolute base on a base url.
The function convert urls of the following tags and attributes:
    tag:a, attribute: href
    tag:img, attribute: src
If you want to add more tags, just set the customTags parameter.
*/
func AbsUrlHtml(baseurl, htmlstr string, customTags ...map[string]string) (h string, err error) {

    base, err := url.Parse(baseurl)
    if err != nil {
        return
    }

    var tags map[string]string

    if len(customTags) > 0 && customTags[0] != nil {
        tags = customTags[0]
    } else {
        // default tags
        tags = make(map[string]string)
        tags["a"] = "href"
        tags["img"] = "src"
    }

    doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(htmlstr)))
    if err != nil {
        return
    }

    for key, value := range tags {
        sel := doc.Find(key)
        for i:= range sel.Nodes {
            single := sel.Eq(i)
            original_url, ok := single.Attr(value)
            if !ok {
                continue
            }

            ourl, err := url.Parse(original_url)
            if err != nil {
                continue
            }

            single.SetAttr(value, base.ResolveReference(ourl).String())
        }
    }

    err = nil
    h, err = doc.Html()
    return
}

