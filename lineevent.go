package dispatcher

import (
	"fmt"
	"reflect"

	"github.com/codegangsta/inject"
)

type LineEvent struct {
	fork      Fork
	listeners map[string]*LineEventChain
}

type chain struct {
	v *reflect.Value
	f int
}
type LineEventChain struct {
	callbacks map[string]*chain
	f         int
}

func newLineEventChain() *LineEventChain {
	return &LineEventChain{callbacks: map[string]*chain{}}
}

type event struct {
	eventName string
	inject.Injector
}

func newEvent(eventName string, args ...interface{}) *event {
	e := event{eventName: eventName, Injector: inject.New()}
	for _, v := range args {
		e.Map(v)
	}
	return &e
}

func (e *event) clone(eventName string) *event {
	return &event{eventName: eventName, Injector: e.Injector}
}

func NewForkEvent(fork Fork) Events {
	return &LineEvent{
		fork:      fork,
		listeners: map[string]*LineEventChain{},
	}
}

func NewLineEvent() Events {
	return NewForkEvent(NewCbs())
}

func (t *LineEvent) GetFork() Fork {
	return t.fork
}

func (t *LineEvent) dispatchEvent(even *event) {
	t.fork.Push(func() {
		if eventChain, ok := t.listeners[even.eventName]; ok {
			if eventChain.f != 0 {
				if eventChain.f > 0 {
					eventChain.f--
				}
			} else {
				for k, item := range eventChain.callbacks {
					switch item.v.Kind() {
					case reflect.Func:
						elem := item.v
						typ := elem.Type()
						args := []reflect.Value{}
						for i := 0; i != typ.NumIn(); i++ {
							arg := even.Get(typ.In(i))
							if !arg.IsValid() {
								arg = reflect.New(typ.In(i)).Elem()
							}
							args = append(args, arg)
						}
						elem.Call(args)
					case reflect.String:
						t.dispatchEvent(even.clone(item.v.String()))
					default:
						if v, ok := item.v.Interface().(Events); ok {
							v.dispatchEvent(even)
						} else {
							t.removeEvent(even.eventName, k)
							continue
						}
					}
					if item.f != 0 {
						item.f--
						if item.f == 0 {
							t.removeEvent(even.eventName, k)
						}
					}
				}
			}
		}
	})
	t.fork.Join()
}

func (t *LineEvent) ForEventEach(eventName string, f func(interface{})) {
	eventChain := t.getChain(eventName)
	for _, item := range eventChain.callbacks {
		f(item.v.Interface())
	}
}

func (t *LineEvent) getChain(eventName string) *LineEventChain {
	eventChain, ok := t.listeners[eventName]
	if !ok {
		eventChain = newLineEventChain()
		t.listeners[eventName] = eventChain
	}
	return eventChain
}

func (t *LineEvent) code(i interface{}) (s string) {
	elem := reflect.ValueOf(i)
	switch elem.Kind() {
	case reflect.String:
		s = "___" + elem.String()
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		s = fmt.Sprint("_", elem.Pointer())
	default:
		s = fmt.Sprintln("__", i)
	}
	return
}

func (t *LineEvent) codes(i interface{}, j []interface{}) string {
	k := t.code(i)
	for _, v := range j {
		k += t.code(v)
	}
	return k
}

func (t *LineEvent) addEvent(eventName string, size int, callback interface{}, token []interface{}) {
	eventChain := t.getChain(eventName)
	elem := reflect.ValueOf(callback)

	k := t.codes(callback, token)

	if eventChain.callbacks[k] != nil {
		return
	}
	eventChain.callbacks[k] = &chain{
		v: &elem,
		f: size,
	}
	return
}

func (t *LineEvent) AddEvent(eventName string, callback interface{}, token ...interface{}) {
	t.addEvent(eventName, 0, callback, token)
}

func (t *LineEvent) OnlyOnce(eventName string, callback interface{}, token ...interface{}) {
	t.addEvent(eventName, 1, callback, token)
}

func (t *LineEvent) OnlyTimes(eventName string, size int, callback interface{}, token ...interface{}) {
	t.addEvent(eventName, size, callback, token)
}

func (t *LineEvent) StopOnce(eventName string) {
	if t.listeners[eventName] != nil {
		t.listeners[eventName].f = 1
	}
}

func (t *LineEvent) IsOpen(eventName string) bool {
	if t.listeners[eventName] != nil {
		return 0 == t.listeners[eventName].f
	}
	return true
}

func (t *LineEvent) CloseEvent(eventName string) {
	if t.listeners[eventName] != nil {
		t.listeners[eventName].f = -1
	}
}

func (t *LineEvent) OpenEvent(eventName string) {
	if t.listeners[eventName] != nil {
		t.listeners[eventName].f = 0
	}
}

func (t *LineEvent) removeEvent(eventName string, index string) {
	if eventChain, ok := t.listeners[eventName]; ok && eventChain.callbacks != nil {
		delete(eventChain.callbacks, index)
	}
}

func (t *LineEvent) RemoveEvent(eventName string, callback interface{}, token ...interface{}) {
	t.removeEvent(eventName, t.codes(callback, token))
}

func (t *LineEvent) Dispatch(eventName string, args ...interface{}) {
	t.dispatchEvent(newEvent(eventName, args...))
}

func (t *LineEvent) Empty() {
	t.listeners = map[string]*LineEventChain{}
}

func (t *LineEvent) EmptyEvent(eventName string) {
	delete(t.listeners, eventName)
}

func (t *LineEvent) Range(eventin, eventout string, events map[string]interface{}, token ...interface{}) {
	t.AddEvent(eventin, func() {
		for k, v := range events {
			t.AddEvent(k, v, token...)
		}
	})
	t.AddEvent(eventout, func() {
		for k, v := range events {
			t.RemoveEvent(k, v, token...)
		}
	})
}
