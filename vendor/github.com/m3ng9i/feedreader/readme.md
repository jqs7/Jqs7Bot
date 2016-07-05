Feedreader
===========

Feedreader is a Go package for parsing RSS 2.0 and Atom 1.0 feed. 

Feedreader on github: <http://github.com/m3ng9i/feedreader>

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

- [RSS 2.0 Specification](http://www.rssboard.org/rss-specification)
- [The Atom Syndication Format](http://tools.ietf.org/html/rfc4287)

## 中文说明

Feedreader包可以解析RSS 2.0与Atom 1.0标准的feed。使用方法可以看上面的例子，然后看一下`Feed`、`FeedItem`和`FeedPerson`的结构。如果要了解完整的功能，可以把包里代码都读一下。

如果RSS或Atom的xml中包含特殊字符，这个包会先将其去除，再进行解析。因此不会出现类似`XML syntax error on line XXX: illegal character code U+XXXX`这样的错误。

因为有些功能我用不上，也就没有实现这些功能：

不支持Atom的`feed/entry/source`节点。

不支持RSS的以下节点：

```
channel/category
channel/docs
channel/cloud
channel/rating
channel/textInput
channel/item/category
channel/item/source
```

我编写这个包参考了下面的规范：

- [RSS 2.0 Specification](http://www.rssboard.org/rss-specification)
- [The Atom Syndication Format](http://tools.ietf.org/html/rfc4287)
