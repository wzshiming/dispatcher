package dispatcher

type Events interface {
	dispatchEvent(event *event)
	AddEvent(eventName string, callback interface{}, token ...interface{}) string
	OnlyOnce(eventName string, callback interface{}, token ...interface{}) string
	OnlyTimes(eventName string, size int, callback interface{}, token ...interface{}) string
	RemoveEvent(eventName string, callback interface{}, token ...interface{})
	RemoveEventIndex(eventName string, index string)
	EventSize(eventName string) int
	Range(eventin, eventout string, events map[string]interface{})
	RangeForOther(e Events, eventin string, eventout string, events map[string]interface{})
	StopOnce(eventName string)
	IsOpen(eventName string) bool
	CloseEvent(eventName string)
	OpenEvent(eventName string)
	Dispatch(eventName string, args ...interface{})
	EmptyEvent(eventName string)
	Empty()
	GetFork() Fork
	ForEventEach(eventName string, f func(string, interface{}))
}

type Fork interface {
	Push(f func())
	Join()
}
