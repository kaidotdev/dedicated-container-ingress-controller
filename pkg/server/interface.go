package server

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

type IProcessor interface {
	Start() error
	Stop(context.Context) error
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
