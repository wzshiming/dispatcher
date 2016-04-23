package dispatcher

import (
	"fmt"
	"reflect"

	"github.com/codegangsta/inject"
	"github.com/wzshiming/base"
	"github.com/wzshiming/slicefunc"
)

type LineEvent struct {
	fork      Fork
	listeners map[string]*LineEventChain
}

type chain struct {
	v *reflect.Value
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
							t.RemoveEventIndex(even.eventName, k)
							continue
						}
					}
				}
			}
		}
	})
	t.fork.Join()
}

func (t *LineEvent) ForEventEach(eventName string, f func(string, interface{})) {
	eventChain := t.getChain(eventName)
	for k, item := range eventChain.callbacks {
		f(k, item.v.Interface())
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

func (t *LineEvent) addEvent(eventName string, callback interface{}) *eventLine {
	eventChain := t.getChain(eventName)
	elem := reflect.ValueOf(callback)
	var k string
	for {
		k = base.RandString()
		if eventChain.callbacks[k] == nil {
			break
		}
	}
	eventChain.callbacks[k] = &chain{
		v: &elem,
	}
	return newEventLine(eventName, k, t, callback)
}

// 添加事件
func (t *LineEvent) AddEvent(eventName string, callback interface{}) *eventLine {
	return t.addEvent(eventName, callback)
}

// 添加事件只执行一次
func (t *LineEvent) OnlyOnce(eventName string, callback interface{}) *eventLine {
	return t.OnlyTimes(eventName, 1, callback)
}

// 添加事件 执行限定次数
func (t *LineEvent) OnlyTimes(eventName string, size int, callback interface{}) *eventLine {
	i := 0
	var e *eventLine
	sf := func() {
		i++
		if i == size {
			e.Close()
		}
	}
	e = t.addEvent(eventName, slicefunc.Join(callback, sf))
	return e
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

func (t *LineEvent) RemoveEventIndex(eventName string, index string) {
	if eventChain, ok := t.listeners[eventName]; ok && eventChain.callbacks != nil {
		delete(eventChain.callbacks, index)
	}
}

func (t *LineEvent) EventSize(eventName string) int {
	if eventChain, ok := t.listeners[eventName]; ok && eventChain.callbacks != nil {
		return len(eventChain.callbacks)
	}
	return 0
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

// 为自己增减操作
func (t *LineEvent) Range(eventin, eventout string, events map[string]interface{}) {
	t.RangeForOther(t, eventin, eventout, events)
}

// 当某个状态时为另一个 事件集合 增减操作
func (t *LineEvent) RangeForOther(e Events, eventin, eventout string, events map[string]interface{}) {
	var el = newEventLines()
	if eventin == "" {
		for k, v := range events {
			el.Append(e.AddEvent(k, v))
		}
		t.OnlyOnce(eventout, func() {
			el.Close()
		})
	} else {
		t.AddEvent(eventin, func() {
			for k, v := range events {
				el.Append(e.AddEvent(k, v))
			}
		})
		t.AddEvent(eventout, func() {
			el.Close()
		})
	}

}
