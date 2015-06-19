package dispatcher

import (
	"testing"
)

type MClass struct {
	dispatcher Dispatcher
	t          *testing.T
}

func Test_tbs(t *testing.T) {
	mc := &MClass{t: t}
	mc.Start()
}

func (t *MClass) Start() {

	dispatcher := NewDispatcher()

	dispatcher.AddEventListener("test", t.onTest)

	dispatcher.AddEventListener("test", t.onTest2)

	dispatcher.Dispatch("test", 10, "hehe")

	dispatcher.RemoveEventListener("test", t.onTest)

	dispatcher.Dispatch("test", "hehe2", 11)

	dispatcher.Empty()

	dispatcher.Dispatch("test", "hehe2", 11)

}

func (t *MClass) onTest(a string) {
	t.t.Log(a)
}

func (t *MClass) onTest2(a string, b int) {
	t.t.Log(a, b)
}
