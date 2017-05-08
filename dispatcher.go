package dispatch

import (
	"fmt"
	"sync"
)

var DispatcherInstanceMutex sync.RWMutex
var ListenerMutex sync.RWMutex

type Dispatcher struct {
	listeners map[string]*EventChain
}

type Event struct {
	eventName string
	Params    map[string]interface{}
}

func CreateEvent(eventName string, params map[string]interface{}) *Event {
	return &Event{eventName: eventName, Params: params}
}

type EventCallback func(*Event)

var _instance *Dispatcher

func SharedDispatcher() *Dispatcher {
	DispatcherInstanceMutex.Lock()
	defer DispatcherInstanceMutex.Unlock()

	if _instance == nil {
		_instance = &Dispatcher{}
		_instance.Init()
	}

	return _instance
}

func (d *Dispatcher) Init() {
	d.listeners = make(map[string]*EventChain)
}

func (d *Dispatcher) safeAddListener(eventName string) *EventChain {
	ListenerMutex.Lock()
	defer ListenerMutex.Unlock()

	eventChain, ok := d.listeners[eventName]
	if !ok {
		eventChain = NewEventChain()
		d.listeners[eventName] = eventChain
	}
	return eventChain
}

func (d *Dispatcher) AddEventListener(eventName string, callback *EventCallback) {
	eventChain := d.safeAddListener(eventName)

	if eventChain.callbackExists(callback) {
		return
	}

	ch := eventChain.addCallback(callback)

	// Start event listener loop for this callback
	go d.handler(eventName, ch, callback)
}

// Event listener loop
//    Blocks on ch until an event comes in. Once an event
//    has been fired, handle the event using the passed-in
//    callback then loop around to wait for the next event
//    to come in
func (d *Dispatcher) handler(eventName string, ch chan *Event, callback *EventCallback) {
	//fmt.Printf("add listener: %s\n", eventName)
	//fmt.Println("chan: ", ch)
	for {
		event := <-ch
		// fmt.Println("event out:", eventName, event, ch)
		if event == nil {
			break
		}
		go (*callback)(event)
	}
}

func (d *Dispatcher) listenersForName(eventName string) (*EventChain, error) {
	ListenerMutex.RLock()
	defer ListenerMutex.RUnlock()

	eventChain, ok := d.listeners[eventName]
	if !ok {
		return nil, fmt.Errorf("Event name '%s' not found", eventName)
	}

	return eventChain, nil
}

func (d *Dispatcher) RemoveEventListener(eventName string, callback *EventCallback) {
	eventChain, err := d.listenersForName(eventName)
	if err != nil {
		return
	}

	eventChain.removeCallback(callback)
}

func (d *Dispatcher) DispatchEvent(event *Event) {
	ListenerMutex.RLock()
	defer ListenerMutex.RUnlock()

	eventChain, ok := d.listeners[event.eventName]
	if ok {
		// fmt.Printf("dispatch event: %s\n", event.eventName)
		eventChain.sendEventToChs(event)
	}
}
