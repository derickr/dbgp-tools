package server

import (
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"io"
	"time"
)

type Logger interface {
	LogInfo(category string, format string, data ...interface{})
	LogWarning(category string, format string, data ...interface{})
	LogError(category string, format string, data ...interface{})
	LogUserInfo(category string, format string, user string, data ...interface{})
	LogUserWarning(category string, format string, user string, data ...interface{})
	LogUserError(category string, format string, user string, data ...interface{})
}

type ConsoleLogger struct {
	output io.Writer
}

func NewConsoleLogger(output io.Writer) *ConsoleLogger {
	return &ConsoleLogger{output: output}
}

func (logger *ConsoleLogger) logHandler(severity Value, category string, user string, format string, data []interface{}) {
	if len(user) > 0 {
		fmt.Fprintf(logger.output, "%s [%s] [%s] [%s] %s\n", time.Now().UTC().Format("15:04:05.000"), severity, White(category), Bold(White(user)), fmt.Sprintf(format, data...))
	} else {
		fmt.Fprintf(logger.output, "%s [%s] [%s] %s\n", time.Now().UTC().Format("15:04:05.000"), severity, White(category), fmt.Sprintf(format, data...))
	}
}

func (logger *ConsoleLogger) LogInfo(category string, format string, data ...interface{}) {
	logger.logHandler(BrightGreen("info"), category, "", format, data)
}

func (logger *ConsoleLogger) LogWarning(category string, format string, data ...interface{}) {
	logger.logHandler(BrightYellow("warn"), category, "", format, data)
}

func (logger *ConsoleLogger) LogError(category string, format string, data ...interface{}) {
	logger.logHandler(BrightRed("err "), category, "", format, data)
}

func (logger *ConsoleLogger) LogUserInfo(category string, user string, format string, data ...interface{}) {
	logger.logHandler(BrightGreen("info"), category, user, format, data)
}

func (logger *ConsoleLogger) LogUserWarning(category string, user string, format string, data ...interface{}) {
	logger.logHandler(BrightYellow("warn"), category, user, format, data)
}

func (logger *ConsoleLogger) LogUserError(category string, user string, format string, data ...interface{}) {
	logger.logHandler(BrightRed("err "), category, user, format, data)
}
