package control

import (
	"encoding/json"
	"testing"
	"time"
)

type TestDelayDurationMessage struct {
	Delay DelayDuration `json:"delay"`
}

func TestDelayDuration(t *testing.T) {
	gotMsg, err := json.Marshal(&TestDelayDurationMessage{
		Delay: DelayDuration{time.Second * 5},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(gotMsg) != `{"delay":"5s"}` {
		t.Errorf("unexpected msg: %s", gotMsg)
	}

	var msg TestDelayDurationMessage
	if err := json.Unmarshal([]byte(`{"delay": "1s"}`), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Delay.Duration != time.Second {
		t.Errorf("unexpected delay: %s", msg.Delay)
	}

	if err := json.Unmarshal([]byte(`{"delay": 1}`), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Delay.Duration != time.Second {
		t.Errorf("unexpected delay: %s", msg.Delay)
	}

	if err := json.Unmarshal([]byte(`{"delay": 0.1}`), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Delay.Duration != 100*time.Millisecond {
		t.Errorf("unexpected delay: %s", msg.Delay)
	}

	if err := json.Unmarshal([]byte(`{"delay": 0}`), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Delay.Duration != 0 {
		t.Errorf("unexpected delay: %s", msg.Delay)
	}

	if err := json.Unmarshal([]byte(`{"delay": null}`), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Delay.Duration != 0 {
		t.Errorf("unexpected delay: %s", msg.Delay)
	}
}
