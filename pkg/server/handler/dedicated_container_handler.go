package handler

import (
	"dedicated-container-ingress-controller/pkg/client"
	"dedicated-container-ingress-controller/pkg/server/core"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"

	v1 "k8s.io/api/core/v1"

	"golang.org/x/xerrors"
)

var (
	NowFunc = time.Now // nolint:gochecknoglobals
)

type DedicatedContainerHandler struct {
	factory   *core.DedicatedContainerFactory
	redisPool IRedisPool
	router    IRouter
	podsLimit int64
}

func NewDedicatedContainerHandler(
	dedicatedContainerFactory *core.DedicatedContainerFactory,
	redisPool IRedisPool,
	router IRouter,
	podsLimit int64,
) *DedicatedContainerHandler {
	h := &DedicatedContainerHandler{
		factory:   dedicatedContainerFactory,
		redisPool: redisPool,
		router:    router,
		podsLimit: podsLimit,
	}
	return h
}

func (h *DedicatedContainerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := client.GetRequestLogger(r.Context())

	var handleErr error
	defer func() {
		if handleErr != nil {
			switch handleErr.(type) {
			case *ClientError:
				logger.Infof("%+v\n", handleErr)

				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			case *ServerError:
				logger.Errorf("%+v\n", handleErr)

				var temporaryError ITemporaryError
				if xerrors.As(handleErr, &temporaryError) && temporaryError.Temporary() {
					http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
					return
				}
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	}()

	key := trimHost(r.Host)

	routes, err := h.router.Get(r, key)
	if err != nil {
		handleErr = NewServerError(r, err)
		return
	}

	var identifier string
	var host string
	reachable := false
	if routes != nil {
		identifier = routes.Identifier
		host = routes.Host
		reachable = checkReachable(host)
	}
	if routes == nil || !reachable {
		count, err := h.podsCount()
		if err != nil {
			handleErr = NewServerError(r, err)
			return
		}
		if count >= h.podsLimit {
			handleErr = NewServerError(r, &temporaryError{"pods limit exceeded"})
			return
		}
		if !h.factory.HasEntry(key) {
			handleErr = NewClientError(r, xerrors.Errorf("factory does not have %s entry", key))
			return
		}
		pod, err := h.factory.Create(r.Context(), key)
		if err != nil {
			handleErr = NewServerError(r, err)
			return
		}
		identifier = fmt.Sprintf("%s/%s", pod.Name, pod.Namespace)
		host = fmt.Sprintf("%s:%d", pod.Status.PodIP, getHTTPPort(pod))
		if err := h.router.Save(r, w, key, &core.Routes{
			Identifier: identifier,
			Host:       host,
		}); err != nil {
			handleErr = NewServerError(r, err)
			return
		}
	}

	httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   host,
	}).ServeHTTP(w, r)

	if err := h.updateTimestamp(identifier); err != nil {
		handleErr = NewServerError(r, err)
		return
	}
}

func (h *DedicatedContainerHandler) podsCount() (int64, error) {
	conn := h.redisPool.Get()
	defer conn.Close()

	count, err := redis.Int64(conn.Do("ZCOUNT", "pods", "-inf", "+inf"))
	if err != nil {
		return 0, xerrors.Errorf("failed to ZCOUNT at redis: %w", err)
	}
	return count, nil
}

func (h *DedicatedContainerHandler) updateTimestamp(identifier string) error {
	conn := h.redisPool.Get()
	defer conn.Close()

	if _, err := conn.Do("ZADD", "pods", NowFunc().Unix(), identifier); err != nil {
		return xerrors.Errorf("failed to ZADD at redis: %w", err)
	}
	return nil
}

func getHTTPPort(pod *v1.Pod) int32 {
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == "http" {
				return port.ContainerPort
			}
		}
	}
	return 80
}

func checkReachable(host string) bool {
	conn, err := net.DialTimeout("tcp", host, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func trimHost(host string) string {
	if idx := strings.IndexByte(host, ':'); idx > 0 {
		host = host[:idx]
	}
	return host
}
