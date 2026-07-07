package realtime

import (
	"sync"
)

// Hub broadcasts zone events to SSE subscribers (in-process).
type Hub struct {
	mu          sync.RWMutex
	subscribers map[uint]map[chan ZoneEvent]struct{}
	global      map[chan ZoneEvent]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uint]map[chan ZoneEvent]struct{}),
		global:      make(map[chan ZoneEvent]struct{}),
	}
}

func (h *Hub) Subscribe(zoneID uint) chan ZoneEvent {
	ch := make(chan ZoneEvent, 8)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subscribers[zoneID] == nil {
		h.subscribers[zoneID] = make(map[chan ZoneEvent]struct{})
	}
	h.subscribers[zoneID][ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(zoneID uint, ch chan ZoneEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs, ok := h.subscribers[zoneID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(h.subscribers, zoneID)
		}
	}
	close(ch)
}

func (h *Hub) SubscribeGlobal() chan ZoneEvent {
	ch := make(chan ZoneEvent, 16)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.global[ch] = struct{}{}
	return ch
}

func (h *Hub) UnsubscribeGlobal(ch chan ZoneEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.global, ch)
	close(ch)
}

func (h *Hub) Publish(event ZoneEvent) {
	h.mu.RLock()
	subs := h.subscribers[event.ZoneID]
	zoneChans := make([]chan ZoneEvent, 0, len(subs))
	for ch := range subs {
		zoneChans = append(zoneChans, ch)
	}
	globalChans := make([]chan ZoneEvent, 0, len(h.global))
	for ch := range h.global {
		globalChans = append(globalChans, ch)
	}
	h.mu.RUnlock()

	for _, ch := range zoneChans {
		select {
		case ch <- event:
		default:
		}
	}
	for _, ch := range globalChans {
		select {
		case ch <- event:
		default:
		}
	}
}
