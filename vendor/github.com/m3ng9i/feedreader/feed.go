package feedreader

import "time"

import "regexp"
import "strings"
import "fmt"
import "net/url"
import h "github.com/m3ng9i/go-utils/html"

type FeedPerson struct {
	Name  string
	Email string
	Uri   string
}

type FeedItem struct {
	Title   string
	Content string // an article's html content, not escaped
	Author  *FeedPerson
	PubDate time.Time
	Updated time.Time
	Link    string // url to the original article
	Guid    string
}

type Feed struct {
	Type        string // rss or atom
	Version     string // version of rss or atom
	Title       string
	Description string // rss's description or atom's subtitle
	Rights      string
	Icon        string // base64 encoded icon image data, not finished yet
	Link        string // url to the website
	FeedLink    string // url of this feed
	Author      *FeedPerson
	Generator   string
	Updated     time.Time
	Items       []*FeedItem
	Guid        string // if type is atom, this value is atom's id. if type is rss or atom's id is empty, Guid is same as FeedLink
}

/* Try to parse time string in different layouts.

The layouts are:
    RFC1123     Mon, 02 Jan 2006 15:04:05 MST
    RFC1123Z    Mon, 02 Jan 2006 15:04:05 -0700
    RFC822      02 Jan 06 15:04 MST
    RFC822Z     02 Jan 06 15:04 -0700
    RFC3339     2006-01-02T15:04:05Z07:00
    RFC3339Nano 2006-01-02T15:04:05.999999999Z07:00

If error occured, the second element of return value is false,
otherwise it's true.
*/
func ParseTime(s string) (time.Time, bool) {
	layout := []string{time.RFC1123, time.RFC1123Z, time.RFC822, time.RFC822Z,
		time.RFC3339, time.RFC3339Nano}

	layout = append(layout, "2006-01-02 15:04:05 -0700")

	s = strings.TrimSpace(s)

	for _, item := range layout {
		t, e := time.Parse(item, s)
		if e == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// determine if a xml string is a rss or atom
func FeedVerifyString(xmldata string) (feedtype, version string) {

	pattern := regexp.MustCompile(`<rss[^>]*?>`)
	sub := pattern.FindStringSubmatch(xmldata)
	if len(sub) > 0 {

		var t struct {
			Version string `xml:"version,attr"`
		}

		unmarshal([]byte(sub[0]), &t)

		feedtype = "rss"
		version = t.Version
		return
	}

	pattern = regexp.MustCompile(`<feed[^>]*?>`)
	sub = pattern.FindStringSubmatch(xmldata)
	if len(sub) > 0 {
		var t struct {
			Xmlns string `xml:"xmlns,attr"`
		}

		unmarshal([]byte(sub[0]), &t)

		if t.Xmlns == "http://www.w3.org/2005/Atom" {
			feedtype = "atom"
			version = "1.0"
			return
		}
	}

	return
}

// determine if a xml string is a rss or atom
func FeedVerify(xmldata []byte) (feedtype, version string) {
	return FeedVerifyString(string(xmldata))
}

// Parse rss data to *Feed structure.
// If returned error is not nil, it will be ParseError.
func rss20ToFeed(xmldata, feedlink string) (feed *Feed, err error) {

	rss, e := Rss20ParseString(xmldata)
	if e != nil {
		err = &ParseError{Err: e}
		return
	}

	feed = &Feed{}
	feed.Type = "rss"
	feed.Version = "2.0"
	feed.Title = rss.Title
	feed.Description = rss.Description
	feed.Rights = rss.Copyright

	if rss.Image != nil {
		feed.Icon = rss.Image.Url
	}

	feed.Link = rss.Link
	feed.FeedLink = feedlink

	author := &FeedPerson{}
	if rss.ManagingEditor != "" {
		author.Name = rss.ManagingEditor
	} else if rss.WebMaster != "" {
		author.Name = rss.WebMaster
	}
	if author.Name != "" {
		feed.Author = author
	}

	feed.Generator = rss.Generator

	if rss.PubDate.Sub(rss.LastBuildDate) > 0 {
		feed.Updated = rss.PubDate
	} else {
		feed.Updated = rss.LastBuildDate
	}

	feed.Guid = feed.FeedLink

	items := make([]*FeedItem, 0)

	for _, i := range rss.Item {
		item := &FeedItem{}
		item.Title = i.Title
		item.PubDate = i.PubDate
		item.Link = i.Link

		if i.Guid != nil {
			item.Guid = i.Guid.Guid
		}
		if item.Guid == "" {
			item.Guid = item.Link
		}

		author := &FeedPerson{}
		author.Name = i.Author
		if author.Name != "" {
			item.Author = author
		}

		item.Content = transformContent(i.Description)
		items = append(items, item)
	}

	if len(items) > 0 {
		feed.Items = items
	}

	return
}

// Parse atom data to *Feed structure
// If returned error is not nil, it will be ParseError.
func atom10ToFeed(xmldata, feedlink string) (feed *Feed, err error) {

	atom, e := Atom10ParseString(xmldata)
	if e != nil {
		err = &ParseError{Err: e}
		return
	}

	feed = &Feed{}
	feed.Type = "atom"
	feed.Version = "1.0"
	feed.Title = atom.Title.String()
	feed.Description = atom.Subtitle.String()
	feed.Rights = atom.Rights.String()
	feed.Updated = atom.Updated

	if atom.Icon != "" {
		feed.Icon = atom.Icon
	} else if atom.Logo != "" {
		feed.Icon = atom.Logo
	}

	if len(atom.Author) > 0 {
		author := &FeedPerson{}
		author.Name = atom.Author[0].Name
		author.Email = atom.Author[0].Email
		author.Uri = atom.Author[0].Uri

		if author.Name != "" {
			feed.Author = author
		}
	}

	if atom.Generator != nil && atom.Generator.Generator != "" {
		uri := atom.Generator.Uri
		if uri != "" {
			uri = fmt.Sprintf(" <%s>", uri)
		}
		version := atom.Generator.Version
		if version != "" {
			version = fmt.Sprintf(" (%s)", version)
		}
		feed.Generator = fmt.Sprintf("%s%s%s ",
			atom.Generator.Generator, uri, version)
	}

	if feedlink != "" {
		feed.FeedLink = feedlink
	}

	for _, i := range atom.Link {
		if i.Rel == "self" && feed.FeedLink == "" {
			feed.FeedLink = i.Href
		} else if i.Rel == "alternate" {
			feed.Link = i.Href
		}
	}

	if feed.Link == "" {
		for _, i := range atom.Link {
			if i.Rel == "via" {
				feed.FeedLink = i.Href
				break
			}
		}
	}

	feed.Guid = atom.Id
	if feed.Guid == "" {
		feed.Guid = feed.FeedLink
	}

	items := make([]*FeedItem, 0)
	for _, i := range atom.Entry {
		item := &FeedItem{}
		item.Title = i.Title.String()
		item.Guid = i.Id
		item.PubDate = i.Published
		item.Updated = i.Updated

		if len(i.Author) > 0 {
			author := &FeedPerson{}
			author.Name = i.Author[0].Name
			author.Email = i.Author[0].Email
			author.Uri = i.Author[0].Uri
			if author.Name != "" {
				item.Author = author
			}
		}

		for _, j := range i.Link {
			if j.Rel == "alternate" {
				item.Link = j.Href
				break
			}
		}
		if item.Link == "" {
			for _, j := range i.Link {
				if j.Rel == "via" {
					item.Link = j.Href
					break
				}
			}
		}

		if item.Guid == "" {
			item.Guid = item.Link
		}

		if i.Content != nil {
			var t Atom10Text
			t.Content = i.Content.Content
			t.Type = i.Content.Type
			item.Content = t.Html()
		} else {
			item.Content = i.Summary.Html()
		}

		items = append(items, item)
	}

	if len(items) > 0 {
		feed.Items = items
	}

	return
}

// Parse rss 2.0 or atom 1.0 to *Feed structure
// If returned error is not nil, it will be ParseError.
func ParseString(xmldata string, feedlink string) (feed *Feed, err error) {

	feedtype, version := FeedVerifyString(xmldata)
	if feedtype == "rss" && version == "2.0" {
		feed, err = rss20ToFeed(xmldata, feedlink)

	} else if feedtype == "atom" && version == "1.0" {
		feed, err = atom10ToFeed(xmldata, feedlink)

	} else {
		err = &ParseError{Err: fmt.Errorf("Request url: %s is not a valid feed.", feedlink)}
		return
	}

	if err != nil {
		return
	}

	trimSpace(&feed.Title)
	trimSpace(&feed.Description)
	trimSpace(&feed.Link)

	var tmp string

	if feed.Link == "" || feed.Link == "/" {
		tmp, err = h.AbsUrl(feedlink, "/")
		if err == nil {
			feed.Link = tmp
		}
		err = nil
	}

	trimSpace(&feed.FeedLink)
	trimSpace(&feed.Guid)

	if feed.Author != nil {
		trimSpace(&feed.Author.Name)
		trimSpace(&feed.Author.Email)
		trimSpace(&feed.Author.Uri)
	}

	for i, _ := range feed.Items {
		if feed.Items[i] != nil {
			trimSpace(&feed.Items[i].Title)

			if feed.Items[i].Author != nil {
				trimSpace(&feed.Items[i].Author.Name)
				trimSpace(&feed.Items[i].Author.Email)
				trimSpace(&feed.Items[i].Author.Uri)
			}

			trimSpace(&feed.Items[i].Link)
			feed.Items[i].Link, err = h.AbsUrl(feedlink, feed.Items[i].Link)
			tmp, err = h.AbsUrl(feedlink, feed.Items[i].Link)
			if err == nil {
				feed.Items[i].Link = tmp
			}

			trimSpace(&feed.Items[i].Guid)
			_, err = url.Parse(feed.Items[i].Guid)
			// feed.Items[i].Guid) is a url
			if err == nil {
				tmp, err = h.AbsUrl(feedlink, feed.Items[i].Guid)
				if err == nil {
					feed.Items[i].Guid = tmp
				}
			}

			tmp, err = h.AbsUrlHtml(feed.Items[i].Link, feed.Items[i].Content)
			if err == nil {
				feed.Items[i].Content = tmp
			}

			err = nil
		}
	}

	return
}

func trimSpace(s *string) {
	if s != nil {
		*s = strings.TrimSpace(*s)
	}
}

// Parse rss 2.0 or atom 1.0 to *Feed structure
// If returned error is not nil, it will be ParseError.
func Parse(xmldata []byte, feedlink string) (*Feed, error) {
	return ParseString(string(xmldata), feedlink)
}
