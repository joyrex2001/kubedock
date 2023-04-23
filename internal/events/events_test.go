package events

import (
	"testing"

	"github.com/joyrex2001/kubedock/internal/server/filter"
)

func TestEvents(t *testing.T) {
	events := New()
	msgid := "1234-5678"
	events.Publish(msgid, Container, Create)
	el, id := events.Subscribe()
	events.Publish(msgid, Container, Start)
	msg := <-el
	if msg.ID != msgid {
		t.Errorf("invalid msg-id %s - expected %s", msg.ID, msgid)
	}
	if msg.Type != Container {
		t.Errorf("invalid type %s - expected %s", msg.Type, Container)
	}
	if msg.Action != Start {
		t.Errorf("invalid type %s - expected %s", msg.Action, Start)
	}
	events.Unsubscribe(id)
	events.Publish(msgid, Container, Die)
}

func TestMatch(t *testing.T) {
	tests := []struct {
		filter string
		msg    Message
		match  bool
	}{
		{
			filter: `{"type":{"image":true}}`,
			msg:    Message{ID: "1234-5678", Type: "image", Action: "pull"},
			match:  true,
		},
		{
			filter: `{"type":{"image":false}}`,
			msg:    Message{ID: "1234-5678", Type: "image", Action: "pull"},
			match:  false,
		},
		{
			filter: `{"type":{"container":true},"container":{"1234-5678":true}}`,
			msg:    Message{ID: "1234-5678", Type: "container", Action: "create"},
			match:  true,
		},
		{
			filter: `{"type":{"container":true},"container":{"1234-5678":true}}`,
			msg:    Message{ID: "5678-1234", Type: "container", Action: "create"},
			match:  false,
		},
	}
	for i, tst := range tests {
		filtr, _ := filter.New(tst.filter)
		if filtr.Match(&tst.msg) != tst.match {
			t.Errorf("failed test %d - unexpected match", i)
		}
	}
}
