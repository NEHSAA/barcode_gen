//+build appengine
package log

import (
	"context"

	"google.golang.org/appengine"
	gaelog "google.golang.org/appengine/log"
)

type gaeLogger struct {
	ctx context.Context
}

func GetLogger(ctx context.Context) Logger {
	if ctx == nil {
		ctx = appengine.NewContext(nil)
	}
	return &gaeLogger{ctx: ctx}
}

func (l *gaeLogger) Debugf(format string, args ...interface{}) {
	gaelog.Debugf(l.ctx, format, args...)
}

func (l *gaeLogger) Infof(format string, args ...interface{}) {
	gaelog.Infof(l.ctx, format, args...)
}

func (l *gaeLogger) Warningf(format string, args ...interface{}) {
	gaelog.Warningf(l.ctx, format, args...)
}

func (l *gaeLogger) Errorf(format string, args ...interface{}) {
	gaelog.Errorf(l.ctx, format, args...)
}
