package server

import (
	"context"
	"dedicated-container-ingress-controller/pkg/client"
	"dedicated-container-ingress-controller/pkg/server/core"
	"dedicated-container-ingress-controller/pkg/server/processor"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"golang.org/x/xerrors"
)

func Run(a *Args) error {
	i := NewInstance()
	logger := client.NewStandardLogger(a.Verbose)
	i.SetLogger(logger)
	i.SetRedisPool(&redis.Pool{
		MaxActive:       10,
		MaxIdle:         10,
		Wait:            true,
		IdleTimeout:     10 * time.Second,
		MaxConnLifetime: 0,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%d", a.RedisHost, a.RedisPort), redis.DialDatabase(a.RedisDB))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < 3*time.Second {
				return nil
			}
			_, err := c.Do("ping")
			return err
		},
	})
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return xerrors.Errorf("could not create kubernetes config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return xerrors.Errorf("could not create kubernetes client: %w", err)
	}
	i.SetClientset(clientset)

	dedicatedContainerFactory := core.NewDedicatedContainerFactory(i.Clientset())

	api, err := processor.NewAPI(processor.APISettings{
		Address:                   a.APIAddress,
		MaxConnections:            a.APIMaxConnections,
		ReUsePort:                 a.ReUsePort,
		KeepAlived:                a.KeepAlived,
		TCPKeepAliveInterval:      time.Duration(a.TCPKeepAliveInterval) * time.Second,
		RedisPool:                 i.RedisPool(),
		Logger:                    i.Logger(),
		DedicatedContainerFactory: dedicatedContainerFactory,
		CookieSecretKey:           a.CookieSecretKey,
		CookieMaxAge:              a.CookieMaxAge,
		PodsLimit:                 a.PodsLimit,
	})
	if err != nil {
		return xerrors.Errorf("failed to create api: %w", err)
	}
	i.AddProcessor(api)

	controller, err := processor.NewController(processor.ControllerSettings{
		Logger:                    i.Logger(),
		DedicatedContainerFactory: dedicatedContainerFactory,
	})
	if err != nil {
		return xerrors.Errorf("failed to create controller: %w", err)
	}
	i.AddProcessor(controller)

	garbageCollector, err := processor.NewGarbageCollector(processor.GarbageCollectorSettings{
		Interval:  time.Minute,
		Lifetime:  time.Hour,
		RedisPool: i.RedisPool(),
		Clientset: i.Clientset(),
		Logger:    i.Logger(),
	})
	if err != nil {
		return xerrors.Errorf("failed to create garbage collector: %w", err)
	}
	i.AddProcessor(garbageCollector)

	monitor, err := processor.NewMonitor(processor.MonitorSettings{
		Address:              a.MonitorAddress,
		MaxConnections:       a.MonitorMaxConnections,
		JaegerEndpoint:       a.MonitoringJaegerEndpoint,
		EnableProfiling:      a.EnableProfiling,
		EnableTracing:        a.EnableTracing,
		TracingSampleRate:    a.TracingSampleRate,
		ReUsePort:            a.ReUsePort,
		KeepAlived:           a.KeepAlived,
		TCPKeepAliveInterval: time.Duration(a.TCPKeepAliveInterval) * time.Second,
		Logger:               i.Logger(),
	})
	if err != nil {
		return xerrors.Errorf("failed to create monitor: %w", err)
	}
	i.AddProcessor(monitor)

	i.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
	i.logger.Infof("Attempt to shutdown instance...\n")

	i.Shutdown(context.Background())
	return nil
}
