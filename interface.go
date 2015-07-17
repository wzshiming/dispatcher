package dispatcher

type Events interface {
	AddEventUnlike(eventName string, callback interface{})
	AddEvent(eventName string, callback interface{})
	RemoveEvent(eventName string, callback interface{})
	OnlyOnce(eventName string, callback interface{})
	StopOnce(eventName string)
	Dispatch(eventName string, args ...interface{})
	Range(eventin, eventout string, events map[string]interface{})
	EmptyEvent(eventName string)
	Empty()
}
