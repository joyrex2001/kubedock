package events

import (
	"testing"
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
