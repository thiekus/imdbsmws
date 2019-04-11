package main

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"os"
)

func isFileExists(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}

func newLog() *logrus.Logger {
	log:= logrus.New()
	// Much better logging
	log.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})
	log.SetOutput(colorable.NewColorableStdout())
	return log
}
