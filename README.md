# go-colorable

[![Build Status](https://github.com/mattn/go-colorable/workflows/test/badge.svg)](https://github.com/mattn/go-colorable/actions?query=workflow%3Atest)
[![Codecov](https://codecov.io/gh/mattn/go-colorable/branch/master/graph/badge.svg)](https://codecov.io/gh/mattn/go-colorable)
[![GoDoc](https://godoc.org/github.com/mattn/go-colorable?status.svg)](http://godoc.org/github.com/mattn/go-colorable)
[![Go Report Card](https://goreportcard.com/badge/mattn/go-colorable)](https://goreportcard.com/report/mattn/go-colorable)

Colorable writer for Windows.

For example, most of logger packages doesn't show colors on Windows. (I know we can do it with ansicon. But I don't want.)
This package is possible to handle escape sequence for ansi color on Windows.

## Too Bad!

![](https://raw.githubusercontent.com/mattn/go-colorable/gh-pages/bad.png)


## So Good!

![](https://raw.githubusercontent.com/mattn/go-colorable/gh-pages/good.png)

## Usage

```go
logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
logrus.SetOutput(colorable.NewColorableStdout())

logrus.Info("succeeded")
logrus.Warn("not correct")
logrus.Error("something error")
logrus.Fatal("panic")
```

You can compile above code on non-Windows OSs.

## Installation

```
$ go get github.com/mattn/go-colorable
```

# License

MIT

# Author

Yasuhiro Matsumoto (a.k.a mattn)
