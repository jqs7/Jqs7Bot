package feedreader

/*
RSS 2.0 parser

The elements below are not supported.
channel/category
channel/docs
channel/cloud
channel/rating
channel/textInput
channel/item/category
channel/item/source
*/

import "time"

import "strings"
import "fmt"
import "github.com/m3ng9i/go-utils/set"

// rss/image element
type Rss20Image struct {
	Url         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Width       uint   `xml:"width"`
	Height      uint   `xml:"height"`
	Description string `xml:"description"`
}

// rss/item/enclosure element
type Rss20ItemEnclosure struct {
	Url    string `xml:"url,attr"`
	Length uint64 `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

// rss/item/guid element
type Rss20ItemGuid struct {
	Guid           string `xml:",chardata"`
	IsPermaLinkRaw string `xml:"isPermaLink,attr"`
	IsPermaLink    bool
}

// rss/item element
type Rss20Item struct {
	Title       string              `xml:"title"`
	Link        string              `xml:"link"`
	Description string              `xml:"description"`
	Author      string              `xml:"author"`
	Comments    string              `xml:"comments"`
	Enclosure   *Rss20ItemEnclosure `xml:"enclosure"`
	Guid        *Rss20ItemGuid      `xml:"guid"`
	PubDateRaw  string              `xml:"pubDate"`
	PubDate     time.Time
}

// whole rss element
type Rss20 struct {
	Version          string `xml:"version,attr"`
	Title            string `xml:"channel>title"`
	Link             string
	Description      string `xml:"channel>description"`
	Language         string `xml:"channel>language"`
	Copyright        string `xml:"channel>copyright"`
	ManagingEditor   string `xml:"channel>managingEditor"`
	WebMaster        string `xml:"channel>webMaster"`
	PubDateRaw       string `xml:"channel>pubDate"`
	PubDate          time.Time
	LastBuildDateRaw string `xml:"channel>lastBuildDate"`
	LastBuildDate    time.Time
	Generator        string      `xml:"channel>generator"`
	Ttl              uint        `xml:"channel>ttl"`
	Image            *Rss20Image `xml:"channel>image"`
	SkipHours        []uint8     `xml:"channel>skipHours>hours"`
	SkipDaysRaw      []string    `xml:"channel>skipDays>days"`
	SkipDays         []time.Weekday
	Item             []*Rss20Item `xml:"channel>item"`
}

func weekdayToNumber(s string) (week time.Weekday, ok bool) {
	ok = true
	switch strings.ToLower(s) {
	case "monday":
		week = time.Monday
		return
	case "tuesday":
		week = time.Tuesday
		return
	case "wednesday":
		week = time.Wednesday
		return
	case "thursday":
		week = time.Thursday
		return
	case "friday":
		week = time.Friday
		return
	case "saturday":
		week = time.Saturday
		return
	case "sunday":
		week = time.Sunday
		return
	}

	ok = false // error
	return
}

// Parse rss xml data to *RSS20 structure.
// If returned error is not nil, it will be ParseError.
func Rss20Parse(b []byte) (rss *Rss20, err error) {

	e := unmarshal(b, &rss)
	if e != nil {
		err = &ParseError{Err: e}
		return
	}

	if rss.Version != "2.0" {
		err = &ParseError{Err: fmt.Errorf("RSS version: %s is not supported.", rss.Version)}
		return
	}

	rss.PubDate, _ = ParseTime(rss.PubDateRaw)
	rss.LastBuildDate, _ = ParseTime(rss.LastBuildDateRaw)

	days := set.New()
	for _, i := range rss.SkipDaysRaw {
		n, ok := weekdayToNumber(i)
		if ok {
			days.Add(n)
		}
	}
	daysArray := make([]time.Weekday, 0, days.Len())
	for _, i := range days.List() {
		v, _ := i.(time.Weekday)
		daysArray = append(daysArray, v)
	}
	rss.SkipDays = daysArray

	for i := range rss.Item {
		if rss.Item[i].Guid != nil {
			rss.Item[i].Guid.IsPermaLink = strings.ToLower(
				rss.Item[i].Guid.IsPermaLinkRaw) == "true"
		}

		rss.Item[i].PubDate, _ = ParseTime(rss.Item[i].PubDateRaw)
	}

	rss.Link, err = rss20ParseLink(b)
	return
}

// Parse rss xml data to *RSS20 structure
// If returned error is not nil, it will be ParseError.
func Rss20ParseString(xmldata string) (*Rss20, error) {
	return Rss20Parse([]byte(xmldata))
}

/*
Parse link element of rss.

The below rss example provide two link elements: one with empty namespace and another with namespace 'atom'.

<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Title</title>
    <description>Some description.</description>
	<link>http://www.example.com/</link>
    <atom:link href="http://www.example.com/feed.xml" rel="self" type="application/rss+xml"/>
    ...
  </channel>
</rss>

If unmarshal the xml into the following struct:
    Link string `xml:"channel>link"`
the second link(the one with namespace)'s value will override the first one, so the Link will be empty string, which is not expected.

This function could solve the problem.
*/
func rss20ParseLink(b []byte) (link string, err error) {
	var t struct {
		Link []string `xml:"channel>link"`
	}

	err = unmarshal(b, &t)
	if err != nil {
		return
	}

	for _, item := range t.Link {
		if item != "" {
			link = item
			return
		}
	}

	return
}
