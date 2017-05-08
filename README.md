# Dispatch

Create listeners and event emitters to allow code to react to events (like named channels).

## How to use it

The dispatcher creates a singleton instance of SharedDispatcher that can be used both to subscribe to events and to emit events.

To create an emitter:
```golang
func emitEvent(obj *Obj) {
	sd := dispatch.SharedDispatcher()
	params := make(map[string]interface{})
	params["obj"] = *obj

	event := sd.CreateEvent("eventName", params)
	sd.DispatchEvent(event)
}
```

To listen for events, you need some simple code:
```golang
type EventListener struct {
	eventAction dispatch.EventCallback
}

func NewEventListener() *EventListener {
	e := new(EventListener)
	e.eventAction = e.onEvent

	sd := dispatch.SharedDispatcher()
  // This "eventName" is the same as the one called from `CreateEvent`
	sd.AddEventListener("eventName", &e.eventAction)

	return e
}

func (e *EventListener) onEvent(event *dispatch.Event) {
	// Perform an action when this is called
}
```
