package server

import (
	"math"
)

type Args struct {
	APIAddress               string
	APIMaxConnections        int64
	MonitorAddress           string
	MonitorMaxConnections    int64
	MonitoringJaegerEndpoint string
	EnableProfiling          bool
	EnableTracing            bool
	TracingSampleRate        float64
	RedisHost                string
	RedisPort                int64
	RedisDB                  int
	CookieSecretKey          string
	CookieMaxAge             int
	PodsLimit                int64
	KeepAlived               bool
	ReUsePort                bool
	TCPKeepAliveInterval     int64
	Verbose                  bool
}

func DefaultArgs() *Args {
	return &Args{
		APIAddress:               "127.0.0.1:8000",
		APIMaxConnections:        math.MaxInt64,
		MonitorAddress:           "127.0.0.1:9090",
		MonitorMaxConnections:    math.MaxInt64,
		MonitoringJaegerEndpoint: "jaeger-agent.istio-system.svc.cluster.local:6831",
		EnableProfiling:          false,
		EnableTracing:            false,
		TracingSampleRate:        0,
		RedisHost:                "127.0.0.1",
		RedisPort:                6379,
		RedisDB:                  0,
		CookieSecretKey:          "",
		CookieMaxAge:             86400,
		PodsLimit:                30,
		KeepAlived:               true,
		ReUsePort:                true,
		TCPKeepAliveInterval:     0,
		Verbose:                  true,
	}
}
