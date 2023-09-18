package main

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var lastLogLevel logrus.Level = logrus.InfoLevel

func getLogger() *logrus.Logger {
	if logger == nil {
		logger = logrus.New()

		// Set the desired log level (e.g., Debug, Info, Warn, Error, Fatal)
		logger.SetLevel(lastLogLevel)

		// Define colors for log levels (if you're using a terminal)
		formatter := &logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		}
		// Attach the formatter to the logger
		logger.SetFormatter(formatter)
	}

	return logger
}
