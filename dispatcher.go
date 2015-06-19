package dispatcher

import (
	//	"fmt"
	"github.com/codegangsta/inject"
	"reflect"
)

type Dispatcher struct {
	listeners map[string]*EventChain
}

type EventChain struct {
	chs       []chan *Event
	callbacks []*reflect.Value
}

func newEventChain() *EventChain {
	return &EventChain{chs: []chan *Event{}, callbacks: []*reflect.Value{}}
}

type Event struct {
	eventName string
	inject.Injector
}

func NewEvent(eventName string, args ...interface{}) *Event {
	e := Event{eventName: eventName, Injector: inject.New()}
	for _, v := range args {
		e.Map(v)
	}
	return &e
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[string]*EventChain),
	}
}

func (t *Dispatcher) AddEventListener(eventName string, callback interface{}) {
	eventChain, ok := t.listeners[eventName]
	if !ok {
		eventChain = newEventChain()
		t.listeners[eventName] = eventChain
	}
	elem := reflect.ValueOf(callback)
	typ := elem.Type()
	if typ.Kind() != reflect.Func {
		return
	}
	for _, item := range eventChain.callbacks {
		if item.Pointer() == elem.Pointer() {
			return
		}
	}

	ch := make(chan *Event, 128)

	eventChain.chs = append(eventChain.chs, ch)
	eventChain.callbacks = append(eventChain.callbacks, &elem)

	go t.handler(eventName, ch, &elem)
}

func (t *Dispatcher) handler(eventName string, ch chan *Event, elem *reflect.Value) {
	typ := elem.Type()
	for event := range ch {
		args := []reflect.Value{}
		for i := 0; i != typ.NumIn(); i++ {
			args = append(args, event.Get(typ.In(i)))
		}
		elem.Call(args)
	}
}

func (t *Dispatcher) RemoveEventListener(eventName string, callback interface{}) {
	b := reflect.ValueOf(callback)
	if b.Type().Kind() != reflect.Func {
		return
	}
	if eventChain, ok := t.listeners[eventName]; ok {
		for k, item := range eventChain.callbacks {
			if item.Pointer() == b.Pointer() {
				close(eventChain.chs[k])
				eventChain.chs = append(eventChain.chs[:k], eventChain.chs[k+1:]...)
				eventChain.callbacks = append(eventChain.callbacks[:k], eventChain.callbacks[k+1:]...)
			}
		}
	}
}

func (t *Dispatcher) DispatchEvent(event *Event) {
	if eventChain, ok := t.listeners[event.eventName]; ok {
		for _, chEvent := range eventChain.chs {
			chEvent <- event
		}
	}
}

func (t *Dispatcher) Dispatch(eventName string, args ...interface{}) {
	t.DispatchEvent(NewEvent(eventName, args...))
}

func (t *Dispatcher) Empty() {
	for k, chain := range t.listeners {
		for i, _ := range chain.callbacks {
			close(chain.chs[i])
		}
		delete(t.listeners, k)
	}
}
