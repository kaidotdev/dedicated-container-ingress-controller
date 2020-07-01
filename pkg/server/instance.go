package server

import (
	"context"
	"dedicated-container-ingress-controller/pkg/client"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gomodule/redigo/redis"

	"k8s.io/client-go/kubernetes"
)

const (
	gracePeriod = 10
)

type Instance struct {
	processors []IProcessor
	clientset  kubernetes.Interface
	redisPool  IRedisPool
	logger     ILogger
}

func NewInstance() *Instance {
	return &Instance{
		redisPool: &redis.Pool{
			MaxActive:       10,
			MaxIdle:         10,
			Wait:            true,
			IdleTimeout:     10 * time.Second,
			MaxConnLifetime: 0,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", ":6379", redis.DialDatabase(0))
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < 3*time.Second {
					return nil
				}
				_, err := c.Do("ping")
				return err
			},
		},
		logger: client.NewDefaultLogger(),
	}
}

func (i *Instance) RedisPool() IRedisPool {
	return i.redisPool
}

func (i *Instance) SetRedisPool(redisPool IRedisPool) {
	i.redisPool = redisPool
}

func (i *Instance) Logger() ILogger {
	return i.logger
}

func (i *Instance) SetLogger(logger ILogger) {
	i.logger = logger
}

func (i *Instance) Clientset() kubernetes.Interface {
	return i.clientset
}

func (i *Instance) SetClientset(clientset kubernetes.Interface) {
	i.clientset = clientset
}

func (i *Instance) AddProcessor(processor IProcessor) {
	i.processors = append(i.processors, processor)
}

func (i *Instance) Start() {
	for _, processor := range i.processors {
		go func(processor IProcessor) {
			defer func() {
				if err := recover(); err != nil {
					i.logger.Errorf("panic: %+v\n", err)
					i.logger.Debugf("%s\n", debug.Stack())
				}
			}()
			if err := processor.Start(); err != nil && err != http.ErrServerClosed {
				i.logger.Errorf("Failed to listen: %s\n", err)
			}
		}(processor)
	}
}

func (i *Instance) Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(gracePeriod)*time.Second)
	defer cancel()
	for _, p := range i.processors {
		if err := p.Stop(ctx); err != nil {
			i.logger.Errorf("Failed to shutdown: %+v\n", err)
		}
	}
	if err := i.redisPool.Close(); err != nil {
		i.logger.Errorf("Failed to close redis pool: %+v\n", err)
	}
	select {
	case <-ctx.Done():
		i.logger.Infof("Instance shutdown timed out in %d seconds\n", gracePeriod)
	default:
	}
	i.logger.Infof("Instance has been shutdown\n")
}
