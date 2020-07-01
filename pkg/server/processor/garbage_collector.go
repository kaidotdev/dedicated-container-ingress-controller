package processor

import (
	"context"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

var (
	NowFunc = time.Now // nolint:gochecknoglobals
)

type GarbageCollectorSettings struct {
	Interval  time.Duration
	Lifetime  time.Duration
	Logger    ILogger
	RedisPool IRedisPool
	Clientset kubernetes.Interface
}

type GarbageCollector struct {
	interval   time.Duration
	lifetime   time.Duration
	cancelFunc context.CancelFunc
	logger     ILogger
	redisPool  IRedisPool
	clientset  kubernetes.Interface
}

func NewGarbageCollector(settings GarbageCollectorSettings) (*GarbageCollector, error) {
	return &GarbageCollector{
		interval:  settings.Interval,
		lifetime:  settings.Lifetime,
		logger:    settings.Logger,
		redisPool: settings.RedisPool,
		clientset: settings.Clientset,
	}, nil
}

func (c *GarbageCollector) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			func() {
				conn := c.redisPool.Get()
				defer conn.Close()

				identifiers, err := redis.Strings(conn.Do("ZRANGEBYSCORE", "pods", "-inf", NowFunc().Unix()-int64(c.lifetime/time.Second)))
				if err != nil {
					c.logger.Errorf("%+v\n", err)
					return
				}
				for _, identifier := range identifiers {
					s := strings.SplitN(identifier, "/", 2)
					pod, namespace := s[0], s[1]
					if err := c.clientset.CoreV1().Pods(namespace).Delete(ctx, pod, metav1.DeleteOptions{}); err != nil {
						c.logger.Errorf("%+v\n", err)
					}
					if _, err := conn.Do("ZREM", "pods", identifier); err != nil {
						c.logger.Errorf("%+v\n", err)
					}
				}
			}()
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *GarbageCollector) Stop(_ context.Context) error {
	c.cancelFunc()
	return nil
}
