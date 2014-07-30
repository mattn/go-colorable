package main

import (
	"github.com/mattn/go-colorable"
	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.SetOutput(colorable.NewColorableStdout())

	logrus.Info("succeeded")
	logrus.Warn("not correct")
	logrus.Error("something error")
	logrus.Fatal("panic")
}
