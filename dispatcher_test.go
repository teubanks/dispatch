package event_dispatcher

import (
	"runtime"
	"testing"
	"time"
)

// Test Object
//   This object allows us to listen for events to be fired
//   and write tests around the events
type Listener struct {
	callbackOneCount int
	callbackTwoCount int
	cbCh1            chan int
	cbCh2            chan int
}

func NewListener() *Listener {
	l := new(Listener)
	l.cbCh1 = make(chan int, 1)
	l.cbCh2 = make(chan int, 1)
	return l
}

func (l *Listener) onTest(event *Event) {
	l.cbCh1 <- 1
}

func (l *Listener) onTest2(event *Event) {
	l.cbCh2 <- 1
}

func TestSharedDispatcher(t *testing.T) {
	dispatcher := SharedDispatcher()
	dispatcher2 := SharedDispatcher()
	if dispatcher != dispatcher2 {
		t.Error("Expected SharedDispatcher to be a singleton")
	}
}

func TestAddEventListener(t *testing.T) {
	t.Run("Simple AddEventListener Calls", func(t *testing.T) {
		dispatcher := SharedDispatcher()
		dispatcher.Init()
		if len(dispatcher.listeners) != 0 {
			t.Errorf("Expected dispatcher initially have 0 listeners. Got '%d' instead", len(dispatcher.listeners))
		}

		l := Listener{}
		var cb1 EventCallback = l.onTest
		dispatcher.AddEventListener("test", &cb1)

		if len(dispatcher.listeners["test"].callbacks) != 1 {
			t.Errorf("Expected dispatcher to have 1 listener for the 'test' event. Got '%d' instead", len(dispatcher.listeners["test"].callbacks))
		}
	})

	t.Run("Multiple AddEventListener calls", func(t *testing.T) {
		dispatcher := SharedDispatcher()
		dispatcher.Init()
		l := NewListener()
		var cb1 EventCallback = l.onTest
		dispatcher.AddEventListener("test", &cb1)
		dispatcher.AddEventListener("test", &cb1)

		if len(dispatcher.listeners["test"].callbacks) != 1 {
			t.Errorf("Expected dispatcher to have 1 listener for the 'test' event. Got '%d' instead", len(dispatcher.listeners["test"].callbacks))
		}
	})
}

func TestRemoveEventListener(t *testing.T) {
	t.Run("Simple RemoveEventListener calls", func(t *testing.T) {
		dispatcher := SharedDispatcher()
		dispatcher.Init()
		if len(dispatcher.listeners) != 0 {
			t.Errorf("Expected dispatcher initially have 0 listeners. Got '%d' instead", len(dispatcher.listeners))
		}

		l := Listener{}
		var cb1 EventCallback = l.onTest
		dispatcher.AddEventListener("test", &cb1)

		if len(dispatcher.listeners["test"].callbacks) != 1 {
			t.Errorf("Expected dispatcher to have 1 listener. Got '%d' instead", len(dispatcher.listeners))
		}

		dispatcher.RemoveEventListener("test", &cb1)
		if len(dispatcher.listeners["test"].callbacks) != 0 {
			t.Errorf("Expected dispatcher to have 0 listeners. Got '%d' instead", len(dispatcher.listeners))
		}
	})

	t.Run("Multiple RemoveEventListener calls", func(t *testing.T) {
		dispatcher := SharedDispatcher()
		dispatcher.Init()
		if len(dispatcher.listeners) != 0 {
			t.Errorf("Expected dispatcher initially have 0 listeners. Got '%d' instead", len(dispatcher.listeners))
		}

		l := Listener{}
		var cb1 EventCallback = l.onTest
		dispatcher.AddEventListener("test", &cb1)
		dispatcher.RemoveEventListener("test", &cb1)
		callbacks := dispatcher.listeners["test"].callbacks
		if len(callbacks) != 0 {
			t.Errorf("Expected dispatcher to have no listeners. Got '%d' instead", len(callbacks))
		}

		dispatcher.RemoveEventListener("test", &cb1)
		callbacks = dispatcher.listeners["test"].callbacks
		if len(callbacks) != 0 {
			t.Errorf("Expected dispatcher to have no listeners. Got '%d' instead", len(callbacks))
		}
	})

	t.Run("Ignores Nonexistent listeners", func(t *testing.T) {
		dispatcher := SharedDispatcher()
		dispatcher.Init()
		l := Listener{}
		var cb1 EventCallback = l.onTest
		dispatcher.RemoveEventListener("test", &cb1)
	})
}

func TestDispatchEvent(t *testing.T) {
	dispatcher := SharedDispatcher()
	dispatcher.Init()
	l := NewListener()
	var cb1 EventCallback = l.onTest
	dispatcher.AddEventListener("test", &cb1)
	var cb2 EventCallback = l.onTest2
	dispatcher.AddEventListener("test", &cb2)

	params := make(map[string]interface{})
	params["id"] = 1000
	event := CreateEvent("test", params)
	dispatcher.DispatchEvent(event)

	start := time.Now()

	l.callbackOneCount += <-l.cbCh1
	l.callbackTwoCount += <-l.cbCh2
	for {
		if l.callbackOneCount > 0 && l.callbackTwoCount > 0 {
			break
		}
		if time.Now().Sub(start) > 1*time.Second {
			t.Fatal("Timed out waiting for Listener to receive updates")
		}
		runtime.Gosched()
	}

	if l.callbackOneCount != 1 {
		t.Errorf("Expected callbackOneCount to be 1. Got '%d' instead", l.callbackOneCount)
	}

	if l.callbackTwoCount != 1 {
		t.Errorf("Expected callbackTwoCount to be 1. Got '%d' instead", l.callbackTwoCount)
	}
}

func BenchmarkDispatchEvent(b *testing.B) {
	dispatcher := SharedDispatcher()
	dispatcher.Init()
	l := NewListener()
	var cb1 EventCallback = l.onTest
	dispatcher.AddEventListener("test", &cb1)
	var cb2 EventCallback = l.onTest2
	dispatcher.AddEventListener("test", &cb2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := make(map[string]interface{})
		params["id"] = 1000
		event := CreateEvent("test", params)
		dispatcher.DispatchEvent(event)
		l.callbackOneCount += <-l.cbCh1
		l.callbackTwoCount += <-l.cbCh2
	}
}
