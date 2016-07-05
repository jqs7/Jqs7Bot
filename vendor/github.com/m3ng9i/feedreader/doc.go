/*
Feedreader is a Go package for parsing RSS 2.0 and Atom 1.0 feed. 

Feedreader on github: http://github.com/m3ng9i/feedreader

Below is an example, it parse a feed, then print feed title, number of items and all the title of items. 

    package main

    import "fmt"
    import "os"
    import "github.com/m3ng9i/feedreader"

    func main() {
        feed, err := feedreader.Fetch("http://example.com/feed.xml")
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        } else {
            fmt.Println("feed title: ", feed.Title)

            fmt.Printf("There are %d item(s) in the feed\n", len(feed.Items))
            for _, i := range(feed.Items) {
                fmt.Println(i.Title)
            }
        }
    }

Function `Fetch` could parse a feed whatever it is RSS or Atom, it return a `Feed` structure.

For more information, just read the code. First you should read `Feed`, `FeedItem` and `FeedPerson` structure. If you want parse RSS or Atom from scratch, you need to read the rest code of the package.

If you want to know more about RSS and Atom, read the specification below:

RSS 2.0 Specification: http://www.rssboard.org/rss-specification

The Atom Syndication Format: http://tools.ietf.org/html/rfc4287
*/
package feedreader
