# go-colorable

Colorable writer for windows

## Too Bad!


## So Good!


## Usage

```go
logrus.SetOutput(colorable.NewColorableWriter())

logrus.Info("succeeded")
logrus.Warn("not correct")
logrus.Error("something error")
logrus.Fatal("panic")
```

## Installation

```
$ go get github.com/mattn/go-colorable
```

# License

MIT

# Author

Yasuhiro Matsumoto (a.k.a mattn)
