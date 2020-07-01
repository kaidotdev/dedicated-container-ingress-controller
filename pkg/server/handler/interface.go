package handler

import (
	"dedicated-container-ingress-controller/pkg/server/core"
	"net/http"

	"github.com/gomodule/redigo/redis"
)

type ITemporaryError interface {
	Error() string
	Temporary() bool
}

type ILogger interface {
	Errorf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

type IRedisPool interface {
	Get() redis.Conn
	Close() error
}

type IRouter interface {
	Get(*http.Request, string) (*core.Routes, error)
	Save(*http.Request, http.ResponseWriter, string, *core.Routes) error
}
