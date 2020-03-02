// +build e2e_test

package e2e_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestHealthz(t *testing.T) {
	getUrlEnv := func(key string) *url.URL {
		val, ok := os.LookupEnv(key)
		if !ok {
			t.Fatalf("must define env: %v", key)
		}
		u, err := url.Parse(val)
		if err != nil {
			t.Fatalf("env %v must contain url: %v", key, err)
		}
		return u
	}

	_ = getUrlEnv("TEST_MOCK_SERVER_BASE")
	controlBase := getUrlEnv("TEST_MOCK_SERVER_CONTROL_BASE")
	controlHealthz := controlBase
	controlHealthz.Path = "/healthz"

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("control not ready: %v", ctx.Err())
		default:
			t.Log(controlHealthz.String())
			res, err := http.Get(controlHealthz.String())
			if err != nil {
				t.Logf("control /healthz error: %v", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if res.StatusCode != http.StatusOK {
				t.Logf("control /healthz invalid status: %#v", res)
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}
		break
	}
}
