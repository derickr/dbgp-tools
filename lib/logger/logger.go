package logger

type Logger interface {
	LogInfo(category string, format string, data ...interface{})
	LogWarning(category string, format string, data ...interface{})
	LogError(category string, format string, data ...interface{})
	LogUserInfo(category string, format string, user string, data ...interface{})
	LogUserWarning(category string, format string, user string, data ...interface{})
	LogUserError(category string, format string, user string, data ...interface{})
}
