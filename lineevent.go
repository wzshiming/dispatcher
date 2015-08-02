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
	callbacks []chain
	f         int
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
						if item.f == 1 {
							t.RemoveEvent(even.eventName, item.v.Interface())
						} else if item.f > 1 {
							eventChain.callbacks[k].f--
						}
					case reflect.String:
						t.dispatchEvent(even.clone(item.v.String()))
					default:
						if v, ok := item.v.Interface().(Events); ok {
							v.dispatchEvent(even)
						} else {
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

func (t *LineEvent) addEvent(eventName string, callback interface{}, size int, same bool) {
	//	t.fork.Push(func() {
	eventChain := t.getChain(eventName)
	elem := reflect.ValueOf(callback)
	typ := elem.Type()

	switch typ.Kind() {
	case reflect.Func:
		if !same {
			for _, item := range eventChain.callbacks {
				if item.v.Kind() == reflect.Func && item.v.Pointer() == elem.Pointer() {
					fmt.Printf("Repeat to added function %v to the same event ", callback)
					return
				}
			}
		}
	case reflect.String:
		if !same {
			for _, item := range eventChain.callbacks {
				if item.v.Kind() == reflect.String && item.v.String() == elem.String() {
					fmt.Printf("Repeat to added String %v to the same event ", callback)
					return
				}
			}
		}

	default:
		if _, ok := callback.(Events); ok {
			for _, item := range eventChain.callbacks {
				if item.v.Kind() == reflect.Interface && item.v.Interface() == elem.Interface() {
					fmt.Printf("Repeat to added Events %v to the same event ", callback)
					return
				}
			}
		} else {
			fmt.Printf("unsustained to added %v to the same event ", callback)
			return
		}
	}
	eventChain.callbacks = append(eventChain.callbacks, chain{
		v: &elem,
		f: size,
	})
	//	})
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

func (t *LineEvent) removeEvent(eventName string, index int) {
	eventChain := t.listeners[eventName]
	if len(eventChain.callbacks) == 0 {
		return
	} else if len(eventChain.callbacks) == 1 {
		t.listeners[eventName] = newLineEventChain()
	} else {
		eventChain.callbacks = append(eventChain.callbacks[:index], eventChain.callbacks[index+1:]...)
	}
	//eventChain.callbacks = append(eventChain.callbacks[:index], eventChain.callbacks[index+1:]...)
}
func (t *LineEvent) RemoveEvent(eventName string, callback interface{}) {
	//	t.fork.Push(func() {
	b := reflect.ValueOf(callback)
	eventChain, ok := t.listeners[eventName]
	if !ok {
		return
	}
	switch b.Kind() {
	case reflect.Func:
		for k, item := range eventChain.callbacks {
			if item.v.Kind() == reflect.Func && item.v.Pointer() == b.Pointer() {
				t.removeEvent(eventName, k)
			}
		}
	case reflect.String:
		for k, item := range eventChain.callbacks {
			if item.v.Kind() == reflect.String && item.v.String() == b.String() {
				t.removeEvent(eventName, k)
			}
		}
	default:
		if _, ok := callback.(Events); ok {
			for k, item := range eventChain.callbacks {
				if item.v.Kind() == reflect.Interface && item.v.Interface() == b.Interface() {
					t.removeEvent(eventName, k)
				}
			}
		}
	}
	//	})
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
