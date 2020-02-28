package control

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/spuf/mockable-server/storage"
)

func TestHandlerJsonRpc1Url(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	w := httptest.NewRecorder()

	handler := NewHandler(storage.NewQueues())
	handler.ServeHTTP(w, r)

	got := w.Result()
	if got.StatusCode != 404 {
		t.Fatalf("unexpected status: %d", got.StatusCode)
	}

	gotBody, _ := ioutil.ReadAll(got.Body)
	_ = got.Body.Close()
	if string(gotBody) != "Not Found\n" {
		t.Fatalf("unexpected body: %s", gotBody)
	}

}

func TestHandlerJsonRpc1Method(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/rpc/1", strings.NewReader(""))
	w := httptest.NewRecorder()

	handler := NewHandler(storage.NewQueues())
	handler.ServeHTTP(w, r)

	got := w.Result()
	if got.StatusCode != 405 {
		t.Fatalf("unexpected status: %d", got.StatusCode)
	}
	allow := got.Header.Get("Allow")
	if allow != "POST" {
		t.Fatalf("unexpected Allow value: %s", allow)
	}

	gotBody, _ := ioutil.ReadAll(got.Body)
	_ = got.Body.Close()
	if string(gotBody) != "Method Not Allowed\n" {
		t.Fatalf("unexpected body: %s", gotBody)
	}
}

func TestHandlerResponses(t *testing.T) {
	for _, tt := range [...]struct {
		name                string
		queuesResponses     []storage.Message
		body                string
		wantBody            string
		wantQueuesResponses []storage.Message
	}{
		{
			name: "Responses.Push invalid",
			body: `{
				"method": "Responses.Push",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": null,
				"error": "validation: status 0 must be in [100; 600)"
			}`,
			wantQueuesResponses: []storage.Message{},
		},

		{
			name: "Responses.Push item",
			body: `{
				"method": "Responses.Push",
				"params": [{
					"status": 200,
					"headers": {
						"Content-Type": "text/plain",
						"Extra-Header": "value"
					},
					"body": "OK"
				}]
			}`,
			wantBody: `{
				"id": null,
				"result": true,
				"error": null
			}`,
			wantQueuesResponses: []storage.Message{
				{
					Headers: http.Header{
						"Content-Type": {"text/plain"},
						"Extra-Header": {"value"},
					},
					Body:     "OK",
					Response: &storage.Response{Status: 200},
				},
			},
		},

		{
			name: "Responses.List empty",
			body: `{
				"method": "Responses.List",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": [],
				"error": null
			}`,
			wantQueuesResponses: []storage.Message{},
		},

		{
			name: "Responses.List items",
			queuesResponses: []storage.Message{
				{
					Headers: http.Header{
						"Content-Type": {"text/plain"},
						"Extra-Header": {"value"},
					},
					Body:     "OK",
					Response: &storage.Response{Status: 200},
				},
				{
					Response: &storage.Response{},
				},
			},
			body: `{
				"method": "Responses.List",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": [
					{
						"status": 200,
						"headers": {"Content-Type": "text/plain","Extra-Header": "value"},
						"body": "OK"
					},
					{
						"status": 0,
						"headers": {},
						"body": ""
					}
				],
				"error": null
			}`,
		},

		{
			name: "Responses.Clear empty",
			body: `{
				"method": "Responses.Clear",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": true,
				"error": null
			}`,
			wantQueuesResponses: []storage.Message{},
		},

		{
			name: "Responses.Clear items",
			queuesResponses: []storage.Message{
				{
					Response: &storage.Response{},
				},
			},
			body: `{
				"method": "Responses.Clear",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": true,
				"error": null
			}`,
			wantQueuesResponses: []storage.Message{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/rpc/1", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			queues := storage.NewQueues()
			if tt.queuesResponses != nil {
				for _, msg := range tt.queuesResponses {
					if err := queues.Responses.PushLast(msg); err != nil {
						t.Fatalf("PushLast: %v", err)
					}
				}
			}

			handler := NewHandler(queues)
			handler.ServeHTTP(w, r)

			got := w.Result()
			if got.StatusCode != 200 {
				t.Fatalf("unexpected response status code: %d", got.StatusCode)
			}
			contentType := got.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("unexpected response Content-Type value: %s", contentType)
			}

			gotBody, err := ioutil.ReadAll(got.Body)
			_ = got.Body.Close()
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}

			var gotBodyObject, wandBodyObject interface{}
			if err := json.Unmarshal(gotBody, &gotBodyObject); err != nil {
				t.Fatalf("response body is invalid json: %s", gotBody)
			}
			if err := json.Unmarshal([]byte(tt.wantBody), &wandBodyObject); err != nil {
				t.Fatalf("test body is invalid json: %s", tt.wantBody)
			}

			if !reflect.DeepEqual(gotBodyObject, wandBodyObject) {
				t.Errorf("response body mismatch:\n got: %#v\nwant: %#v", gotBodyObject, wandBodyObject)
			}

			if tt.wantQueuesResponses != nil {
				list := queues.Responses.List()
				if !reflect.DeepEqual(list, tt.wantQueuesResponses) {
					t.Errorf("queues.Responses mismatch:\n got: %#v\nwant: %#v", list, tt.wantQueuesResponses)
				}
			}
		})
	}
}

func TestHandlerRequests(t *testing.T) {
	for _, tt := range [...]struct {
		name               string
		queuesRequests     []storage.Message
		body               string
		wantBody           string
		wantQueuesRequests []storage.Message
	}{
		{
			name: "Requests.Pop empty",
			body: `{
				"method": "Requests.Pop",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": null,
				"error": null
			}`,
			wantQueuesRequests: []storage.Message{},
		},

		{
			name: "Requests.Pop item",
			queuesRequests: []storage.Message{
				{
					Headers: http.Header{
						"Content-Type": {"text/plain"},
						"Extra-Header": {"value"},
					},
					Body: "OK",
					Request: &storage.Request{
						Method: "GET",
						Url:    "/path?query",
					},
				},
			},
			body: `{
				"method": "Requests.Pop",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": {
					"method": "GET",
					"url": "/path?query",
					"headers": {"Content-Type": "text/plain","Extra-Header": "value"},
					"body": "OK"
				},
				"error": null
			}`,
			wantQueuesRequests: []storage.Message{},
		},

		{
			name: "Requests.List empty",
			body: `{
				"method": "Requests.List",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": [],
				"error": null
			}`,
			wantQueuesRequests: []storage.Message{},
		},

		{
			name: "Requests.List items",
			queuesRequests: []storage.Message{
				{
					Headers: http.Header{
						"Content-Type": {"text/plain"},
						"Extra-Header": {"value"},
					},
					Body: "OK",
					Request: &storage.Request{
						Method: "GET",
						Url:    "/path?query",
					},
				},
				{
					Request: &storage.Request{},
				},
			},
			body: `{
				"method": "Requests.List",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": [ 
					{
						"method": "GET",
						"url": "/path?query",
						"headers": {"Content-Type": "text/plain","Extra-Header": "value"},
						"body": "OK"
					},
					{
						"method": "",
						"url": "",
						"headers": {},
						"body": ""
					}
				],
				"error": null
			}`,
		},

		{
			name: "Requests.Clear items",
			queuesRequests: []storage.Message{
				{
					Request: &storage.Request{},
				},
			},
			body: `{
				"method": "Requests.Clear",
				"params": []
			}`,
			wantBody: `{
				"id": null,
				"result": true,
				"error": null
			}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/rpc/1", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			queues := storage.NewQueues()
			if tt.queuesRequests != nil {
				for _, msg := range tt.queuesRequests {
					if err := queues.Requests.PushLast(msg); err != nil {
						t.Fatalf("PushLast: %v", err)
					}
				}
			}

			handler := NewHandler(queues)
			handler.ServeHTTP(w, r)

			got := w.Result()
			if got.StatusCode != 200 {
				t.Fatalf("unexpected response status code: %d", got.StatusCode)
			}
			contentType := got.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("unexpected response Content-Type value: %s", contentType)
			}

			gotBody, err := ioutil.ReadAll(got.Body)
			_ = got.Body.Close()
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}

			var gotBodyObject, wandBodyObject interface{}
			if err := json.Unmarshal(gotBody, &gotBodyObject); err != nil {
				t.Fatalf("response body is invalid json: %s", gotBody)
			}
			if err := json.Unmarshal([]byte(tt.wantBody), &wandBodyObject); err != nil {
				t.Fatalf("test body is invalid json: %s", tt.wantBody)
			}

			if !reflect.DeepEqual(gotBodyObject, wandBodyObject) {
				t.Errorf("response body mismatch:\n got: %#v\nwant: %#v", gotBodyObject, wandBodyObject)
			}

			if tt.wantQueuesRequests != nil {
				list := queues.Requests.List()
				if !reflect.DeepEqual(list, tt.wantQueuesRequests) {
					t.Errorf("queues.Requests mismatch:\n got: %#v\nwant: %#v", list, tt.wantQueuesRequests)
				}
			}
		})
	}
}
