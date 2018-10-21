package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type logrusLogger struct {
	log *logrus.Entry
}

func InitLogrus() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "15:04:05.000",
		ForceColors:     true,
		FullTimestamp:   true,
		ForceFormatting: true,
	})
}

func GetLogrusLogger(name string) Logger {

	with_fields := make(logrus.Fields)
	with_fields["prefix"] = fmt.Sprintf("%15s", name)

	internal_log := logrus.WithFields(with_fields)

	return &logrusLogger{
		log: internal_log,
	}
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

func (l *logrusLogger) Warningf(format string, args ...interface{}) {
	l.log.Warningf(format, args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.log.Errorf(format, args...)
}
