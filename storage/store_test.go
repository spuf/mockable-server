package storage

import (
	"strconv"
	"testing"
)

func TestNewStoreEmpty(t *testing.T) {
	store := NewStore()
	list := store.List()
	if len(list) != 0 {
		t.Fatalf("%v must be empty", list)
	}
}

func TestEmptyStoreNoPop(t *testing.T) {
	store := NewStore()
	msg := store.PopFirst()
	if msg != nil {
		t.Fatalf("%v must be nil", msg)
	}
}

func TestStorePush(t *testing.T) {
	store := NewStore()
	body := "body"
	store.PushLast(Message{Body: body})

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("%v must contain one item", list)
	}
	if list[0].Body != body {
		t.Fatalf("%v .Body must be equal to %v", list[0], body)
	}
}

func TestStoreClear(t *testing.T) {
	store := NewStore()
	store.PushLast(Message{})

	store.Clear()

	list := store.List()
	if len(list) != 0 {
		t.Fatalf("%v must be empty", list)
	}
}

func TestStorePop(t *testing.T) {
	store := NewStore()
	body := "body"
	store.PushLast(Message{Body: body})

	msg := store.PopFirst()
	if msg == nil {
		t.Fatalf("%v must not be nil", msg)
	}
	if msg.Body != body {
		t.Fatalf("%v .Body must be equal to %v", msg, body)
	}
}

func TestStorePopNil(t *testing.T) {
	store := NewStore()
	store.PushLast(Message{})

	store.PopFirst()

	msg := store.PopFirst()
	if msg != nil {
		t.Fatalf("%v must be nil", msg)
	}
}

func TestStoreImmutability(t *testing.T) {
	store := NewStore()
	origBody := "origBody"
	origMsg := Message{Body: origBody}
	store.PushLast(origMsg)

	list := store.List()

	list0Body := "list[0].Body"
	list[0].Body = list0Body
	if origMsg.Body != origBody {
		t.Fatalf("%v .Body must be equal to %v", origMsg, origBody)
	}

	msg := store.PopFirst()

	msgBody := "msg.Body"
	msg.Body = msgBody
	if origMsg.Body != origBody {
		t.Fatalf("%v .Body must be equal to %v", origMsg, origBody)
	}

	origMsg.Body = "origMsg.Body"
	if list[0].Body != list0Body {
		t.Fatalf("%v .Body must be equal to %v", list[0], list0Body)
	}
	if msg.Body != msgBody {
		t.Fatalf("%v .Body must be equal to %v", msg, msgBody)
	}
}

func TestStoreLIFO(t *testing.T) {
	store := NewStore()

	for i := 0; i < 5; i++ {
		store.PushLast(Message{Body: strconv.Itoa(i)})
	}

	list := store.List()
	for i, msg := range list {
		if msg.Body != strconv.Itoa(i) {
			t.Fatalf("%v .Body must be %d", msg, i)
		}
	}

	for i := 0; i < 5; i++ {
		msg := store.PopFirst()
		if msg.Body != strconv.Itoa(i) {
			t.Fatalf("%v .Body must be %d", msg, i)
		}
	}
}
