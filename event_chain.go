package dispatch

import (
	"sync"
	"unsafe"
)

type EventChain struct {
	chs       []chan *Event
	callbacks []*EventCallback
	mutex     sync.RWMutex
}

func NewEventChain() *EventChain {
	return &EventChain{chs: []chan *Event{}, callbacks: []*EventCallback{}}
}

func (e *EventChain) addCallback(cb *EventCallback) chan *Event {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	ch := make(chan *Event)
	e.chs = append(e.chs[:], ch)
	e.callbacks = append(e.callbacks[:], cb)
	return ch
}

func (e *EventChain) removeCallback(cb *EventCallback) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	var ch chan *Event
	exist := false
	idx := 0
	for i, item := range e.callbacks {
		a := *(*int)(unsafe.Pointer(item))
		b := *(*int)(unsafe.Pointer(cb))
		if a == b {
			exist = true
			ch = e.chs[i]
			idx = i
			break
		}
	}

	if exist {
		ch <- nil

		e.chs = append(e.chs[:idx], e.chs[idx+1:]...)
		e.callbacks = append(e.callbacks[:idx], e.callbacks[idx+1:]...)
	}
}

func (e *EventChain) callbackExists(cb *EventCallback) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	exist := false
	for _, item := range e.callbacks {
		a := *(*int)(unsafe.Pointer(item))
		b := *(*int)(unsafe.Pointer(cb))
		if a == b {
			exist = true
			break
		}
	}
	return exist
}

func (e *EventChain) sendEventToChs(event *Event) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	for _, chEvent := range e.chs {
		chEvent <- event
	}
}
