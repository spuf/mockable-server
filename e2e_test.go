package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/spuf/mockable-server/control"
	"github.com/spuf/mockable-server/mock"
	"github.com/spuf/mockable-server/storage"
)

func jsonRpcCall(t *testing.T, handler http.Handler, request, expectedResponse string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/rpc/1", strings.NewReader(request))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != 200 {
		t.Errorf("unexpected status: %v", res.StatusCode)
	}
	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("unexpected Content-Type value: %s", contentType)
	}

	actualResponse, _ := ioutil.ReadAll(res.Body)

	var actualResponseObject interface{}
	if err := json.Unmarshal(actualResponse, &actualResponseObject); err != nil {
		t.Fatalf("actual response body is invalid json: %v\n%v", actualResponse, err)
	}
	var expectedResponseObject interface{}
	if err := json.Unmarshal([]byte(expectedResponse), &expectedResponseObject); err != nil {
		t.Fatalf("expected response body is invalid json: %v\n%v", expectedResponse, err)
	}
	if !reflect.DeepEqual(actualResponseObject, expectedResponseObject) {
		t.Errorf("unexpected response: %v", actualResponse)
	}
}

func TestE2E(t *testing.T) {
	// @todo: use table testing like net/http/httptest/httptest_test.go:18

	queues := storage.NewQueues()

	controlHandler := control.NewHandler(queues)
	mockHandler := mock.NewHandler(queues)

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Responses.Push",
			"params": [{
				"status": 200,
				"headers": {
					"Content-Type": "text/plain",
					"Extra-Header": "value"
				},
				"body": "OK"
			}]    
		}      
	`, `{"id":null,"result":true,"error":null}`)

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Responses.List",
			"params": []    
		}      
	`, `{"id":null,"result":[{"status":200,"headers":{"Content-Type":"text/plain","Extra-Header":"value"},"body":"OK"}],"error":null}`)

	{
		req := httptest.NewRequest("PUT", "/path", strings.NewReader(`data`))
		req.Header.Set("Content-Type", "raw/data")
		w := httptest.NewRecorder()

		mockHandler.ServeHTTP(w, req)

		res := w.Result()
		if res.StatusCode != 200 {
			t.Errorf("unexpected status: %v", res.StatusCode)
		}
		contentType := res.Header.Get("Content-Type")
		if contentType != "text/plain" {
			t.Errorf("unexpected Content-Type value: %v", contentType)
		}
		extraHeader := res.Header.Get("Extra-Header")
		if extraHeader != "value" {
			t.Errorf("unexpected Extra-Header value: %v", extraHeader)
		}

		resBody, _ := ioutil.ReadAll(res.Body)

		if string(resBody) != "OK" {
			t.Errorf("unexpected body: %v", resBody)
		}
	}

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Requests.List",
			"params": []    
		}      
	`, `{"id":null,"result":[{"method":"PUT","url":"/path","headers":{"Content-Type":"raw/data"},"body":"data"}],"error":null}`)

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Requests.Pop",
			"params": []    
		}      
	`, `{"id":null,"result":{"method":"PUT","url":"/path","headers":{"Content-Type":"raw/data"},"body":"data"},"error":null}`)
}
