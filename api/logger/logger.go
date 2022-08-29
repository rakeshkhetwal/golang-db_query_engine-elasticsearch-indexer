
package logger

import (
  "time"

  "github.com/sirupsen/logrus" 
)

// Event stores messages to log later, from our standard interface
type Event struct {
  id      int
  message string
}

// StandardLogger enforces specific log message formats
type StandardLogger struct {
  *logrus.Logger
}

// Logger initializes the standard logger
func Logger() *StandardLogger {

  logrus.SetReportCaller(true)
	var baseLogger = logrus.New()
	var standardLogger = &StandardLogger{baseLogger}
	standardLogger.Formatter = &logrus.JSONFormatter{}
  
	return standardLogger
}

// Declare variables to store log messages as new Events
var (
  errorMessage           = Event{1, "%s %v"}
  successMessage         = Event{2, "%s %s"}
  infoMessage            = Event{3, "%s %v"}
  httplogging            = Event{4, "request completed"}
  biStatsErrorMessage    = Event{5, "Indexed [%s] documents with [%s] errors in %v (%s docs/sec)"}
  biStatsSuccessMessage  = Event{6, "Sucessfuly indexed [%s] documents in %v (%s docs/sec)"}
)

func (l *StandardLogger) ErrorMessage(message error, argument string) {
  l.Errorf(errorMessage.message, argument, message)
}

func (l *StandardLogger) BiStatsErrorMessage(documents, errors string, time time.Duration, rate string) {
  l.Errorf(biStatsErrorMessage.message, documents, errors, time, rate)
}

func (l *StandardLogger) FatalErrorMessage(message error, argument string) {
  l.Fatalf(errorMessage.message, argument, message)
}

func (l *StandardLogger) SuccessMessage(message string, argument string) {
  l.Infof(successMessage.message, message, argument)
}

func (l *StandardLogger) BiStatsSuccessMessage(documents string, time time.Duration, rate string) {
  l.Infof(biStatsSuccessMessage.message, documents, time, rate)
}

func (l *StandardLogger) InfoMessage(message string, argument string) {
  l.Infof(infoMessage.message, message, argument)
}

func (l *StandardLogger) HttpLogging(RequestURI, Method string, status int, duration time.Duration, size int) {
  logrus.SetFormatter(&logrus.JSONFormatter{})
  logrus.WithFields(
    logrus.Fields{
        "URI": RequestURI,
        "Method": Method,
        "Status": status,
        "duration": duration,
        "size": size,
    },
  ).Infof(httplogging.message)
}

