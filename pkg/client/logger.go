package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var NowFunc = time.Now // nolint:gochecknoglobals

type ILogger interface {
	Errorf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

type DefaultLogger struct{}

func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

func (dl *DefaultLogger) Errorf(format string, v ...interface{}) {}

func (dl *DefaultLogger) Infof(format string, v ...interface{}) {}

func (dl *DefaultLogger) Debugf(format string, v ...interface{}) {}

type StandardLogger struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger
}

func NewStandardLogger(verbose bool) *StandardLogger {
	logger := &StandardLogger{}
	logger.errorLogger = log.New(os.Stderr, "", log.LUTC)
	logger.infoLogger = log.New(os.Stdout, "", log.LUTC)
	if verbose {
		logger.debugLogger = log.New(os.Stdout, "", log.LUTC)
	} else {
		logger.debugLogger = log.New(ioutil.Discard, "", log.LUTC)
	}
	return logger
}

func (sl *StandardLogger) Errorf(format string, v ...interface{}) {
	sl.errorLogger.Printf(format, v...)
}

func (sl *StandardLogger) Infof(format string, v ...interface{}) {
	sl.infoLogger.Printf(format, v...)
}

func (sl *StandardLogger) Debugf(format string, v ...interface{}) {
	sl.debugLogger.Printf(format, v...)
}

type contextKey string

const requestLoggerContextKey contextKey = "requestLoggerKey"

type RequestLogger struct {
	requestid string
	inner     ILogger
}

func NewRequestLogger(requestid string, logger ILogger) *RequestLogger {
	return &RequestLogger{
		requestid: requestid,
		inner:     logger,
	}
}

func (l *RequestLogger) Errorf(format string, v ...interface{}) {
	l.inner.Errorf(`{"time":"%s","level":"error","requestid":"%s","payload":%q}`, NowFunc().UTC().Format(time.RFC3339Nano), l.requestid, fmt.Sprintf(format, v...))
}

func (l *RequestLogger) Infof(format string, v ...interface{}) {
	l.inner.Infof(`{"time":"%s","level":"info","requestid":"%s","payload":%q}`, NowFunc().UTC().Format(time.RFC3339Nano), l.requestid, fmt.Sprintf(format, v...))
}

func (l *RequestLogger) Debugf(format string, v ...interface{}) {
	l.inner.Debugf(`{"time":"%s","level":"debug","requestid":"%s","payload":%q}`, NowFunc().UTC().Format(time.RFC3339Nano), l.requestid, fmt.Sprintf(format, v...))
}

func GetRequestLogger(ctx context.Context) *RequestLogger {
	v := ctx.Value(requestLoggerContextKey)

	requestLogger, ok := v.(*RequestLogger)
	if !ok {
		return &RequestLogger{
			inner: NewDefaultLogger(),
		}
	}

	return requestLogger
}

func SetRequestLogger(ctx context.Context, requestLogger *RequestLogger) context.Context {
	return context.WithValue(ctx, requestLoggerContextKey, requestLogger)
}
