package feedreader

/*
Atom 1.0 parser
Only support feed documents, not support entry documents.

The elements below are not supported.
    feed/entry/source
*/

import "fmt"
import "time"
import "html"


// Person Constructs: author, contributor element
type Atom10Person struct {
    Name        string  `xml:"name"`        // required
    Email       string  `xml:"email"`
    Uri         string  `xml:"uri"`
}

// Text constructs: rights, title, subtitle, summary
// Type is "xhtml", "html" or "text"
type Atom10Text struct {
    Content     string  `xml:",chardata"`
    Type        string `xml:"type,attr"`
}

// Attribute's of category element
type Atom10Category struct {
    Term        string  `xml:"term,attr"`   // required
    Scheme      string  `xml:"scheme,attr"`
    Label       string  `xml:"label,attr"`
}

// Value and attributes of generator element
type Atom10Generator struct {
    Generator   string  `xml:",chardata"`
    Uri         string  `xml:"uri,attr"`
    Version     string  `xml:"version,attr"`
}

// Attributes of link element
type Atom10Link struct {
    Href        string  `xml:"href,attr"`   // required
    Rel         string  `xml:"rel,attr"`
    Type        string  `xml:"type,attr"`
    Hreflang    string  `xml:"hreflang,attr"`
    Title       string  `xml:"title,attr"`
    Length      uint64  `xml:"length,attr"`
}

type Atom10Content struct {
    Atom10Text
    Src     string  `xml:"src,attr"`
}

// Entry element, not support source element
type Atom10Entry struct {
    Title           *Atom10Text         `xml:"title"`       // required
    Content         *Atom10Content      `xml:"content"`
    Author          []*Atom10Person     `xml:"author"`
    Category        []*Atom10Category   `xml:"category"`
    Contributor     []*Atom10Person     `xml:"contributor"`
    Id              string              `xml:"id"`          // required
    Link            []*Atom10Link       `xml:"link"`
    PublishedRaw    string              `xml:"published"`
    Published       time.Time
    Rights          *Atom10Text         `xml:"rights"`
    Summary         *Atom10Text         `xml:"summary"`
    UpdatedRaw      string              `xml:"updated"`     // required
    Updated         time.Time
}

// Feed element
type Atom10Feed struct {
    Xmlns           string              `xml:"xmlns,attr"`  // feed's attribute
    Author          []*Atom10Person     `xml:"author"`
    Category        []*Atom10Category   `xml:"category"`
    Contributor     []*Atom10Person     `xml:"contributor"`
    Generator       *Atom10Generator    `xml:"generator"`
    Icon            string              `xml:"icon"`
    Logo            string              `xml:"logo"`
    Id              string              `xml:"id"`          // required
    Link            []*Atom10Link       `xml:"link"`
    Rights          *Atom10Text         `xml:"rights"`
    Title           *Atom10Text         `xml:"title"`       // required
    Subtitle        *Atom10Text         `xml:"subtitle"`
    UpdatedRaw      string              `xml:"updated"`     // required
    Updated         time.Time
    Entry           []*Atom10Entry      `xml:"entry"`       // required
}


func (t *Atom10Text) String() string {
    if t == nil {
        return ""
    }

    if t.Type == "xhtml" {

        var inner struct {
            Content string `xml:",innerxml"`
        }

        err := unmarshal([]byte(t.Content), &inner)
        if err != nil {
            return ""
        }

        return html.EscapeString(inner.Content)

    } else {
        return t.Content
    }
}

func (t *Atom10Text) Html() string {
    if t == nil {
        return ""
    }

    if t.Type == "xhtml" {
        var inner struct {
            Content string `xml:",innerxml"`
        }

        err := unmarshal([]byte(t.Content), &inner)
        if err != nil {
            return ""
        }

        return inner.Content

    } else {
        return transformContent(t.Content)
    }
}


/* Parse atom feed document

arguments:
    b       bytes of atom xml
    strict  strict mode or not

return value:
    feed    point to a Atom10Feed struct
    err     error message which type is ParseError
*/
func Atom10Parse(b []byte) (feed *Atom10Feed, err error) {

    const xmlns = "http://www.w3.org/2005/Atom"

    e := unmarshal(b, &feed)
    if e != nil {
        err = &ParseError{Err: e}
        return
    }

    if feed.Xmlns != xmlns {
        err = &ParseError{Err: fmt.Errorf("Atom 1.0's xmlns should be '%s', '%s' is not correct.", xmlns, feed.Xmlns)}
        return
    }

    feed.Updated, _ = ParseTime(feed.UpdatedRaw)

    for i := range(feed.Link) {
        if feed.Link[i].Rel == "" {
            feed.Link[i].Rel = "alternate"
        }
    }

    for i := range(feed.Entry) {

        if feed.Entry[i].Link != nil {
            for j := range(feed.Entry[i].Link) {
                if feed.Entry[i].Link[j].Rel == "" {
                    feed.Entry[i].Link[j].Rel = "alternate"
                }
            }
        }

        feed.Entry[i].Published, _ = ParseTime(feed.Entry[i].PublishedRaw)
        feed.Entry[i].Updated, _ = ParseTime(feed.Entry[i].UpdatedRaw)
    }

    return
}


func Atom10ParseString(s string) (*Atom10Feed, error) {
    return Atom10Parse([]byte(s))
}
