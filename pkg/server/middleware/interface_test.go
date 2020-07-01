package middleware_test

import (
	"dedicated-container-ingress-controller/pkg/client"
	"dedicated-container-ingress-controller/pkg/server/middleware"
	"time"
)

type loggerMock struct {
	middleware.ILogger
	fakeErrorf func(format string, v ...interface{})
	fakeInfof  func(format string, v ...interface{})
	fakeDebugf func(format string, v ...interface{})
}

func (l loggerMock) Errorf(format string, v ...interface{}) {
	l.fakeErrorf(format, v...)
}

func (l loggerMock) Infof(format string, v ...interface{}) {
	l.fakeInfof(format, v...)
}

func (l loggerMock) Debugf(format string, v ...interface{}) {
	l.fakeDebugf(format, v...)
}

func init() {
	client.NowFunc = func() time.Time {
		return time.Unix(0, 0)
	}
}