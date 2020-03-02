// +build e2e_test

package e2e_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type tt struct {
	name        string
	request     http.Request
	want        http.Response
	compareJson bool
}

func newTextTest(name string, method, url, body string, wantStatus int, wantBody string) tt {
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	return tt{
		name:    name,
		request: *request,
		want: http.Response{
			StatusCode: wantStatus,
			Body:       ioutil.NopCloser(strings.NewReader(wantBody)),
		},
		compareJson: false,
	}
}

func newJsonRpcTest(name string, url, body string, wantBody string) tt {
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	return tt{
		name:    name,
		request: *request,
		want: http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(wantBody)),
		},
		compareJson: true,
	}
}

func TestE2E(t *testing.T) {
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

	mockBase := getUrlEnv("TEST_MOCK_SERVER_BASE").String()

	controlBase := getUrlEnv("TEST_MOCK_SERVER_CONTROL_BASE")
	controlBase.Path = "/healthz"
	controlHealthz := controlBase.String()
	controlBase.Path = "/rpc/1"
	controlJsonRpc := controlBase.String()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("control not ready: %v", ctx.Err())
		default:
			res, err := http.Get(controlHealthz)
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

	for _, tt := range [...]tt{
		newJsonRpcTest("clear requests", controlJsonRpc, `{
			"method": "Requests.Clear",
			"params": []
		}`, `{
			"id": null,
			"result": true,
			"error": null
		}`),
		newJsonRpcTest("clear responses", controlJsonRpc, `{
			"method": "Responses.Clear",
			"params": []
		}`, `{
			"id": null,
			"result": true,
			"error": null
		}`),
		newJsonRpcTest("push response", controlJsonRpc, `{
			"method": "Responses.Push",
			"params": [{
				"status": 201,
				"headers": {
					"Content-Type": "text/plain",
					"Extra-Header": "value"
				},
				"body": "Hello"
			}]
		}`, `{
			"id": null,
			"result": true,
			"error": null
		}`),
		newTextTest("request mock server", http.MethodGet, mockBase, "", 201, "Hello"),
		newJsonRpcTest("pop request", controlJsonRpc, `{
			"method": "Requests.Pop",
			"params": []
		}`, `{
			"id": null,
			"result": {
				"method": "GET",
				"url": "/",
				"headers": {"Accept-Encoding": "gzip","User-Agent": "Go-http-client/1.1"},
				"body": ""
			},
			"error": null
		}`),
		newJsonRpcTest("list requests empty", controlJsonRpc, `{
			"method": "Requests.List",
			"params": []
		}`, `{
			"id": null,
			"result": [],
			"error": null
		}`),
		newJsonRpcTest("list responses empty", controlJsonRpc, `{
			"method": "Responses.List",
			"params": []
		}`, `{
			"id": null,
			"result": [],
			"error": null
		}`),
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{
				Timeout: time.Minute,
			}
			got, err := client.Do(&tt.request)
			if err != nil {
				t.Fatalf("client.Do: %v", err)
			}

			if got.StatusCode != tt.want.StatusCode {
				t.Errorf("status code mismatch:\n got: %v\nwant: %v", got.StatusCode, tt.want.StatusCode)
			}

			gotBody, err := ioutil.ReadAll(got.Body)
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}
			wantBody, err := ioutil.ReadAll(tt.want.Body)
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}

			if tt.compareJson {
				var gotBodyObject, wandBodyObject interface{}
				if err := json.Unmarshal(gotBody, &gotBodyObject); err != nil {
					t.Fatalf("response body is invalid json: %v\n%v", gotBody, err)
				}
				if err := json.Unmarshal(wantBody, &wandBodyObject); err != nil {
					t.Fatalf("test body is invalid json: %v\n%v", wantBody, err)
				}

				if !reflect.DeepEqual(gotBodyObject, wandBodyObject) {
					t.Errorf("response json mismatch:\n got: %#v\nwant: %#v", gotBodyObject, wandBodyObject)
				}
			} else {
				if string(gotBody) != string(wantBody) {
					t.Errorf("response body mismatch:\n got: %#v\nwant: %#v", string(gotBody), string(wantBody))
				}
			}
		})
	}
}
