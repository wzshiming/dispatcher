package dispatcher

type eventLine struct {
	eventName string
	token     string
	parent    Events
	callback  interface{}
}

func newEventLine(eventName string, token string, parent Events, callback interface{}) *eventLine {
	return &eventLine{
		eventName: eventName,
		token:     token,
		parent:    parent,
		callback:  callback,
	}
}

func (t *eventLine) Name() string {
	return t.eventName
}

func (t *eventLine) Close() {
	t.parent.RemoveEventIndex(t.eventName, t.token)
}

type eventLines []*eventLine

func newEventLines() *eventLines {
	return &eventLines{}
}

func (t *eventLines) Append(el ...*eventLine) {
	*t = append(*t, el...)
}

func (t *eventLines) Close() {
	for _, v := range *t {
		v.Close()
	}
}
