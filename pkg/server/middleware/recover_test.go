package middleware_test

import (
	"context"
	"dedicated-container-ingress-controller/pkg/client"
	"dedicated-container-ingress-controller/pkg/server/middleware"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		receiver http.Handler
		in       *http.Request
	}{
		{
			func() string {
				_, _, line, _ := runtime.Caller(1)
				return fmt.Sprintf("L%d", line)
			}(),
			middleware.NewRecoverMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			httptest.NewRequest("GET", "/", nil).WithContext(client.SetRequestLogger(context.Background(), client.NewRequestLogger("", loggerMock{}))),
		},
		{
			func() string {
				_, _, line, _ := runtime.Caller(1)
				return fmt.Sprintf("L%d", line)
			}(),
			middleware.NewRecoverMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(errors.New("fake"))
			})),
			httptest.NewRequest("GET", "/", nil).WithContext(client.SetRequestLogger(context.Background(), client.NewRequestLogger("", loggerMock{
				fakeErrorf: func(format string, v ...interface{}) {
					stack := "panic: fake\n"
					want := fmt.Sprintf(`{"time":"1970-01-01T00:00:00Z","level":"error","requestid":"","payload":%q}`, stack)
					got := fmt.Sprintf(format, v...)
					if diff := cmp.Diff(want, got); diff != "" {
						t.Errorf("(-want +got):\n%s", diff)
					}
				},
				fakeDebugf: func(format string, v ...interface{}) {},
			}))),
		},
	}
	for _, tt := range tests {
		got := httptest.NewRecorder()

		name := tt.name
		receiver := tt.receiver
		in := tt.in
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			receiver.ServeHTTP(got, in)
		})
	}
}
