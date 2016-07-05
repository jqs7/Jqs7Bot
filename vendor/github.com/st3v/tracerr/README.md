tracerr
=======

[![Build Status](https://travis-ci.org/st3v/tracerr.svg?branch=master)](https://travis-ci.org/st3v/tracerr)

[![GoDoc](https://godoc.org/github.com/st3v/tracerr?status.png)](http://godoc.org/github.com/st3v/tracerr)

Traceable errors in Go.

#### Example:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/st3v/tracerr"
)

func main() {
	foo := &foo{}

	err := nested(4, func() error {
		return foo.bar()
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}

func nested(depth int, fn func() error) error {
	if depth <= 1 {
		return fn()
	}
	return nested(depth-1, fn)
}

type foo struct{}

func (f *foo) bar() error {
	return tracerr.Wrap(errors.New("FooBarError"))
}
```

#### Output:

```
$ ./example
FooBarError
  at (*foo).bar (main/tracerr_example.go:32)
  at func·001 (main/tracerr_example.go:14)
  at nested (main/tracerr_example.go:24)
  at nested (main/tracerr_example.go:26)
  at nested (main/tracerr_example.go:26)
  at nested (main/tracerr_example.go:26)
  at main (main/tracerr_example.go:15)
  at main (runtime/proc.c:255)
  at goexit (runtime/proc.c:1445)
```
