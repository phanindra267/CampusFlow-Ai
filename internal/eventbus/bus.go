package eventbus

import (
	"log"
	"sync"
)

type Event struct {
	Type    string
	Payload interface{}
}

type EventBus struct {
	subscribers map[string][]chan Event
	rm          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (b *EventBus) Subscribe(eventType string, ch chan Event) {
	b.rm.Lock()
	defer b.rm.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
}

func (b *EventBus) Publish(event Event) {
	b.rm.RLock()
	defer b.rm.RUnlock()
	if chans, found := b.subscribers[event.Type]; found {
		for _, ch := range chans {
			select {
			case ch <- event:
			default:
				log.Printf("EventBus: subscriber channel for %s is full, dropping event", event.Type)
			}
		}
	}
}