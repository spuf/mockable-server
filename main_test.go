package main

import (
	"bytes"
	"encoding/json"
	"log"
	"mockable-server/storage"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func jsonEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

func TestFunctional(t *testing.T) {
	queues := &Queues{
		Responses: storage.NewStore(),
		Requests:  storage.NewStore(),
	}

	controlServer := httptest.NewServer(newControlHandler(queues))
	defer controlServer.Close()

	mockServer := httptest.NewServer(newMockHandle(queues))
	defer mockServer.Close()

	if res, err := http.Post(controlServer.URL+"/rpc/1", "application/json", strings.NewReader(`
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
	`)); err != nil {
		log.Fatal(err)
	} else {
		var body bytes.Buffer
		if _, err := body.ReadFrom(res.Body); err != nil {
			t.Fatal(err)
		}
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}

		if same, err := jsonEqual(body.Bytes(), []byte(`{"id":null,"result":true,"error":null}`)); err != nil {
			t.Fatal(err)
		} else if !same {
			t.Fatalf("unexpected response: %s", body.String())
		}
	}
}
