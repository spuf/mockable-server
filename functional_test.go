package main

import (
	"encoding/json"
	"github.com/spuf/mockable-server/control"
	"github.com/spuf/mockable-server/mock"
	"github.com/spuf/mockable-server/storage"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func jsonRpcCall(t *testing.T, handler http.Handler, request, expectedResponse string) {
	t.Helper()

	req := httptest.NewRequest("POST", "/rpc/1", strings.NewReader(request))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != 200 {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("unexpected Content-Type value: %s", contentType)
	}

	actualResponse, _ := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()

	var actualResponseObject interface{}
	if err := json.Unmarshal(actualResponse, &actualResponseObject); err != nil {
		t.Fatalf("actual response body is invalid json: %s", actualResponse)
	}
	var expectedResponseObject interface{}
	if err := json.Unmarshal([]byte(expectedResponse), &expectedResponseObject); err != nil {
		t.Fatalf("expected response body is invalid json: %s", expectedResponse)
	}
	if !reflect.DeepEqual(actualResponseObject, expectedResponseObject) {
		t.Fatalf("unexpected response: %s", actualResponse)
	}
}

func TestFunctional(t *testing.T) {
	// @todo: use table testing like net/http/httptest/httptest_test.go:18

	queues := storage.NewQueues()

	controlHandler := control.NewHandler(queues)
	mockHandler := mock.NewHandler(queues)

	{
		req := httptest.NewRequest("GET", "/", strings.NewReader(""))
		w := httptest.NewRecorder()

		controlHandler.ServeHTTP(w, req)

		res := w.Result()
		if res.StatusCode != 404 {
			t.Fatalf("unexpected status: %d", res.StatusCode)
		}

		resBody, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if string(resBody) != "Not Found\n" {
			t.Fatalf("unexpected body: %s", resBody)
		}
	}

	{
		req := httptest.NewRequest("GET", "/rpc/1", strings.NewReader(""))
		w := httptest.NewRecorder()

		controlHandler.ServeHTTP(w, req)

		res := w.Result()
		if res.StatusCode != 405 {
			t.Fatalf("unexpected status: %d", res.StatusCode)
		}
		allow := res.Header.Get("Allow")
		if allow != "POST" {
			t.Fatalf("unexpected allow value: %s", allow)
		}

		resBody, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if string(resBody) != "Method Not Allowed\n" {
			t.Fatalf("unexpected body: %s", resBody)
		}
	}

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Responses.Push",
			"params": []    
		}      
	`, `{"id":null,"result":null,"error":"validation: status 0 must be in [100; 600)"}`)

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
			t.Fatalf("unexpected status: %d", res.StatusCode)
		}
		contentType := res.Header.Get("Content-Type")
		if contentType != "text/plain" {
			t.Fatalf("unexpected Content-Type value: %s", contentType)
		}
		extraHeader := res.Header.Get("Extra-Header")
		if extraHeader != "value" {
			t.Fatalf("unexpected Extra-Header value: %s", extraHeader)
		}

		resBody, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		if string(resBody) != "OK" {
			t.Fatalf("unexpected body: %s", resBody)
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

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Requests.Pop",
			"params": []    
		}      
	`, `{"id":null,"result":null,"error":null}`)

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Responses.Clear",
			"params": []    
		}      
	`, `{"id":null,"result":true,"error":null}`)

	jsonRpcCall(t, controlHandler, `
		{
			"method": "Requests.Clear",
			"params": []    
		}      
	`, `{"id":null,"result":true,"error":null}`)
}
