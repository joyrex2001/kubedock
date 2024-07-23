package events

import (
	"sync"
	"time"

	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/util/stringid"
)

// Events is the interface to publish and consume events.
type Events interface {
	Subscribe() (<-chan Message, string)
	Unsubscribe(string)
	Publish(string, string, string)
}

// instance is the internal representation of the Events object.
type instance struct {
	mu        sync.Mutex
	observers map[string]chan Message
}

var singleton *instance
var once sync.Once

// New will create return the singleton Events instance.
func New() Events {
	once.Do(func() {
		singleton = &instance{}
		singleton.observers = map[string]chan Message{}
	})
	return singleton
}

// Publish will publish an event for given resource id and type for given action.
func (e *instance) Publish(id, typ, action string) {
	msg := Message{ID: id, Type: typ, Action: action}
	msg.Time = time.Now().Unix()
	msg.TimeNano = time.Now().UnixNano()
	for _, ob := range e.observers {
		ob <- msg
	}
}

// Subscribe will subscribe to the events and will return a channel and an
// unique identifier than can be used to unsubscribe when done.
func (e *instance) Subscribe() (<-chan Message, string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make(chan Message, 1)
	id := stringid.GenerateRandomID()
	e.observers[id] = out
	klog.V(5).Infof("subscribing %s to events", id)
	return out, id
}

// Unsubscribe will unsubscribe given subscriber id from the events.
func (e *instance) Unsubscribe(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	klog.V(5).Infof("unsubscribing %s from events", id)
	delete(e.observers, id)
}

// Match will match given event filter conditions.
func (m *Message) Match(typ string, key string, val string) bool {
	klog.V(5).Infof("match %s: %s = %s", typ, key, val)
	if typ == Type {
		return m.Type == key
	}
	if m.Type == typ {
		return m.ID == key
	}
	return true
}
