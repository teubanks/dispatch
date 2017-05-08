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
		// fmt.Println("remove", a, b)
		if a == b {
			exist = true
			ch = e.chs[i]
			idx = i
			break
		}
	}

	if exist {
		// fmt.Printf("remove listener: %s\n", eventName)
		// fmt.Println("chan: ", ch)
		ch <- nil

		e.chs = append(e.chs[:idx], e.chs[idx+1:]...)
		e.callbacks = append(e.callbacks[:idx], e.callbacks[idx+1:]...)
		// fmt.Println(len(eventChain.chs))
	}
}

func (e *EventChain) callbackExists(cb *EventCallback) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	exist := false
	//fmt.Println("add len:", len(eventChain.callbacks))
	for _, item := range e.callbacks {
		a := *(*int)(unsafe.Pointer(item))
		b := *(*int)(unsafe.Pointer(cb))
		//fmt.Println("add", a, b)
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
