# gomtch - find text even if it doesn't want to be found

Do your users have clever ways to hide some terms from you? Sometimes it is hard to find 
forbidden terms when the user doesn't want it to be found. 

![technology Go](https://img.shields.io/badge/technology-go-blue.svg) [![Build Status](https://travis-ci.com/nicolasassi/gomtch.svg?branch=main)](https://travis-ci.com/nicolasassi/gomtch) [![GoDoc](https://godoc.org/github.com/nicolasassi/gomtch?status.svg)](https://pkg.go.dev/github.com/nicolasassi/gomtch) [![GoReportCard](https://goreportcard.com/badge/github.com/nicolasassi/gomtch)](https://goreportcard.com/report/nicolasassi/gomtch)

gomtch aims to help you find tokens in real life text offering the flexibility that most
out-of-the-box algorithms lack.
Ever wanted to find instances of a split word in text corpora (s p l i t e d)? Most NLP algorithms
require a lot of normalization what could warm the integrity of the text corpora you are working
with. gomtch looks for instances of splited words making the whole process easier. Also, the
classic duplicated character problem (reeeeal), gomtch takes care of that for you as well.
Finally, gomtch gives you the possibility to choose how to analise a potentially dangerous text
corpora by considering special characters and digits as wild cards and leaving to you to choose
how much (%) of a term should be considered (ex: h4rd matches 90% with the word hard).

https://nicolasassi.medium.com/gomtch-find-text-even-if-it-doesnt-want-to-be-found-a2229aed2a88


## Table of Contents

* [Installation](#installation)
* [Docs](#docs)
* [API](#api)
* [Examples](#examples)
* [Support](#support)
* [License](#license)

## Installation

just your good old `go get`

    $ go get github.com/nicolasassi/gomtch

(optional) To run unit tests:

    $ cd $GOPATH/src/github.com/nicolasassi/gomtch
    $ go test

(optional) To run benchmarks (warning: it might take some time):

    $ cd $GOPATH/src/github.com/nicolasassi/gomtch
    $ go test -bench=".*"


## Docs

https://pkg.go.dev/github.com/nicolasassi/gomtch

## API

gomtch exposes a Document interface. The porpoise of it `Document` is to be compared with another `Document` interface.
Keep in mind that one `Document` can be compared with as many `Documents` as necessary. Use the `Scan()` in the
reference `Document` with the `Documents` do be compared as arguments (ex: referenceDocument.Scan(doc1, doc2 doc3...)).

gomtch provides a variety of text normalization features. Some features already implemented are:

- HTML parsing (remove any HTML tags and keep the text)
- Sequential character removal (reaaal = real)
- Upper and lower normalization
- Unicode normalization (canção = cancao)
- Replace any unwanted token with a regexp

The implementation of `Document` requires the field matchScoreFunc with has the following signature 
`func(int, int) bool`. This field is used to determine the percentage of a token that should match a token.

Note that the matching behaves differently when comparing digits, letters and special characters.

## Examples

### Simple document

```Go 
package main

import (
    "bytes"
    "github.com/nicolasassi/gomtch"
    "log"
)

func main() {
    text := []byte("this is a text c o r p o r a")
    tokenToFind := []byte("corpora")
    corp, err := gomtch.NewDoc(bytes.NewReader(text))
    if err != nil {
        log.Fatal(err)
    }
    match, err := gomtch.NewDoc(bytes.NewReader(tokenToFind))
    if err != nil {
        log.Fatal(err)
    }
    for index, match := range corp.Scan(match) {
        log.Printf("index: %v match: %s", index, string(match))
    }
}
```

### Playing with matching scores

```Go 
package main

import (
  "bytes"
  "github.com/nicolasassi/gomtch"
  "log"
)

func main() {
  text := []byte("this is a text corp0ra")
  tokenToFind := []byte("corpora")
  corp, err := gomtch.NewDoc(bytes.NewReader(text))
  if err != nil {
    log.Fatal(err)
  }
  // this will not match because the default minimum match score of NewDoc is 100 and
  // "corp0ra" != "corpora"
  match1, err := gomtch.NewDoc(bytes.NewReader(tokenToFind))
  if err != nil {
    log.Fatal(err)
  }
  // this will match because 90% of len(tokenToFind) == +- 6. This means that there is space for
  // one not matching letter.
  match2, err := gomtch.NewDoc(bytes.NewReader(tokenToFind), gomtch.WithMinimumMatchScore(90))
  if err != nil {
    log.Fatal(err)
  }
  for index, match := range corp.Scan(match1, match2) {
    log.Printf("index: %v match: %s", index, string(match))
  }
}
```

### Complex document and conditions

```Go 
package main

import (
  "bytes"
  "github.com/nicolasassi/gomtch"
  "log"
  "regexp"
)

func main() {
    text := []byte("<p>This is REAAAAAAL WORLD example of a téxt quite h4rd to match!!<p>")
    corp, err := gomtch.NewDoc(bytes.NewReader(text),
        gomtch.WithSetLower(),
        gomtch.WithSequentialEqualCharsRemoval(),
        gomtch.WithHMTLParsing(),
        gomtch.WithReplacer(regexp.MustCompile(`[\[\]()\-.,:;{}"'!?]`), " "),
        gomtch.WithTransform(gomtch.NewASCII()))
    if err != nil {
        log.Fatal(err)
    }
    // this will match because we set that sequential equal caracters shoud removed and the
    // text should be all in lower case.
    // So REAAAAAL becames real
    match1, err := gomtch.NewDoc(bytes.NewReader([]byte("real world")))
    if err != nil {
        log.Fatal(err)
    }
    // this will match because we allow each token to have a 90% minimum match score.
    match2, err := gomtch.NewDoc(bytes.NewReader([]byte("text quite hard to match")),
        gomtch.WithMinimumMatchScore(90))
    if err != nil {
        log.Fatal(err)
    }
    for index, match := range corp.Scan(match1, match2) {
        log.Printf("index: %v match: %s", index, string(match))
    }
}
```

## Support

There are a number of ways you can support the project:

* Use it, star it, build something with it, spread the word!
* Raise issues to improve the project (note: doc typos and clarifications are issues too!)
    - Please search existing issues before opening a new one - it may have already been addressed.
* Pull requests: please discuss new code in an issue first, unless the fix is really trivial.
    - Make sure new code is tested.
    - Be mindful of existing code - PRs that break existing code have a high probability of being declined, unless it fixes a serious issue.

## License

The BSD 3-Clause license, the same as the Go language.
