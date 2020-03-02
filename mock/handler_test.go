package mock

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/spuf/mockable-server/storage"
)

func TestHandlerNoResponse(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/base/../path?query", strings.NewReader("Hello"))
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	queues := storage.NewQueues()
	handler := NewHandler(queues)
	handler.ServeHTTP(w, r)

	got := w.Result()
	if got.StatusCode != 501 {
		t.Errorf("unexpected status code: %v", got.StatusCode)
	}
	gotBody, _ := ioutil.ReadAll(got.Body)
	if string(gotBody) != "Not Implemented\n" {
		t.Errorf("unexpected body: %v", string(gotBody))
	}

	msg := queues.Requests.PopFirst()
	want := &storage.Message{
		Headers: http.Header{
			"Content-Type": {"text/plain"},
		},
		Body: "Hello",
		Request: &storage.Request{
			Method: "POST",
			Url:    "/base/../path?query",
		},
	}

	if !reflect.DeepEqual(msg, want) {
		t.Errorf("mismatch request:\n got: %#v\nwant:%#v", msg, want)
	}
}

func TestHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/base/../path?query", strings.NewReader("Hello"))
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	queues := storage.NewQueues()
	res := storage.Message{
		Headers: http.Header{
			"Content-Type": {"text/plain"},
		},
		Body:     "Answer",
		Request:  nil,
		Response: &storage.Response{Status: 201},
	}
	if err := queues.Responses.PushLast(res); err != nil {
		t.Fatalf("PushLast: %v", err)
	}

	handler := NewHandler(queues)
	handler.ServeHTTP(w, r)

	got := w.Result()
	if got.StatusCode != 201 {
		t.Errorf("unexpected status code: %v", got.StatusCode)
	}
	gotContentType := got.Header.Get("Content-Type")
	if gotContentType != "text/plain" {
		t.Errorf("unexpected status code: %v", gotContentType)
	}
	gotBody, _ := ioutil.ReadAll(got.Body)
	if string(gotBody) != "Answer" {
		t.Errorf("unexpected body: %v", string(gotBody))
	}

	msg := queues.Requests.PopFirst()
	want := &storage.Message{
		Headers: http.Header{
			"Content-Type": {"text/plain"},
		},
		Body: "Hello",
		Request: &storage.Request{
			Method: "POST",
			Url:    "/base/../path?query",
		},
	}

	if !reflect.DeepEqual(msg, want) {
		t.Errorf("mismatch request:\n got: %#v\nwant:%#v", msg, want)
	}
}
