package dispatcher

type Events interface {
	dispatchEvent(event *event)
	AddEventUnlike(eventName string, callback interface{})
	AddEvent(eventName string, callback interface{})
	RemoveEvent(eventName string, callback interface{})
	OnlyOnce(eventName string, callback interface{})
	StopOnce(eventName string)
	CloseEvent(eventName string)
	OpenEvent(eventName string)
	Dispatch(eventName string, args ...interface{})
	Range(eventin, eventout string, events map[string]interface{})
	EmptyEvent(eventName string)
	Empty()
	GetFork() Fork
	ForEventEach(eventName string, f func(interface{}))
}

type Fork interface {
	Push(f func())
	Join()
}
