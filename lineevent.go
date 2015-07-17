package dispatcher

import (
	"fmt"
	"reflect"

	"github.com/codegangsta/inject"
)

type LineEvent struct {
	listeners map[string]*LineEventChain
}

type chain struct {
	v *reflect.Value
	f uint
}
type LineEventChain struct {
	callbacks []chain
	f         uint
}

func newLineEventChain() *LineEventChain {
	return &LineEventChain{callbacks: []chain{}}
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

func (t *LineEvent) dispatchEvent(event *event) {
	if eventChain, ok := t.listeners[event.eventName]; ok {
		if eventChain.f != 0 {
			eventChain.f--
		} else {
			for k, item := range eventChain.callbacks {
				switch item.v.Kind() {
				case reflect.Func:
					elem := item.v
					typ := elem.Type()
					args := []reflect.Value{}
					for i := 0; i != typ.NumIn(); i++ {
						arg := event.Get(typ.In(i))
						if !arg.IsValid() {
							arg = reflect.New(typ.In(i)).Elem()
						}
						args = append(args, arg)
					}
					elem.Call(args)
					if item.f == 1 {
						t.RemoveEvent(event.eventName, item.v.Interface())
					} else if item.f > 1 {
						eventChain.callbacks[k].f--
					}
				case reflect.String:
					t.dispatchEvent(event.clone(item.v.String()))
				}
			}
		}
	}
}

func NewLineEvent() *LineEvent {
	return &LineEvent{
		listeners: make(map[string]*LineEventChain),
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

func (t *LineEvent) addEvent(eventName string, callback interface{}, size uint, same bool) {
	eventChain := t.getChain(eventName)
	elem := reflect.ValueOf(callback)
	typ := elem.Type()
	if !same {
		switch typ.Kind() {
		case reflect.Func:
			for _, item := range eventChain.callbacks {
				if item.v.Kind() == reflect.Func && item.v.Pointer() == elem.Pointer() {
					return
				}
			}
		case reflect.String:
			for _, item := range eventChain.callbacks {
				if item.v.Kind() == reflect.String && item.v.String() == elem.String() {
					return
				}
			}
		default:
			fmt.Printf("Repeat to added function %v to the same event ", callback)
			return
		}
	}
	eventChain.callbacks = append(eventChain.callbacks, chain{
		v: &elem,
		f: size,
	})
	return
}

func (t *LineEvent) AddEventUnlike(eventName string, callback interface{}) {
	t.addEvent(eventName, callback, 0, false)
}

func (t *LineEvent) AddEvent(eventName string, callback interface{}) {
	t.addEvent(eventName, callback, 0, true)
}

func (t *LineEvent) OnlyOnce(eventName string, callback interface{}) {
	t.addEvent(eventName, callback, 1, true)
}

func (t *LineEvent) StopOnce(eventName string) {
	if t.listeners[eventName] != nil {
		t.listeners[eventName].f = 1
	}
}

func (t *LineEvent) RemoveEvent(eventName string, callback interface{}) {
	b := reflect.ValueOf(callback)
	eventChain, ok := t.listeners[eventName]
	if !ok {
		return
	}
	switch b.Kind() {
	case reflect.Func:
		for k, item := range eventChain.callbacks {
			if item.v.Kind() == reflect.Func && item.v.Pointer() == b.Pointer() {
				eventChain.callbacks = append(eventChain.callbacks[:k], eventChain.callbacks[k+1:]...)
			}
		}
	case reflect.String:
		for k, item := range eventChain.callbacks {
			if item.v.Kind() == reflect.String && item.v.String() == b.String() {
				eventChain.callbacks = append(eventChain.callbacks[:k], eventChain.callbacks[k+1:]...)
			}
		}
	}
}

func (t *LineEvent) Dispatch(eventName string, args ...interface{}) {
	t.dispatchEvent(newEvent(eventName, args...))
}

func (t *LineEvent) Empty() {
	t.listeners = make(map[string]*LineEventChain)
}

func (t *LineEvent) EmptyEvent(eventName string) {
	delete(t.listeners, eventName)
}

func (t *LineEvent) Range(eventin, eventout string, events map[string]interface{}) {
	t.AddEvent(eventin, func() {
		for k, v := range events {
			t.AddEvent(k, v)
		}
	})
	t.AddEvent(eventout, func() {
		for k, v := range events {
			t.RemoveEvent(k, v)
		}
	})
}
