package dispatcher

import (
	"testing"
)

type MClass struct {
	dispatcher Events
	t          *testing.T
}

func Test_tbs(t *testing.T) {
	mc := &MClass{t: t}
	mc.Start()
}

func (t *MClass) Start() {

	dispatcher := NewLineEvent()

	dispatcher.OnlyOnce("test", t.onTest)

	dispatcher.OnlyOnce("test2", t.onTest2)
	dispatcher.OnlyOnce("test", "test2")
	dispatcher.StopOnce("test")

	dispatcher.Dispatch("test", 10, "hehe")

	dispatcher.Dispatch("test", "hehe2", 11)
	dispatcher.Dispatch("test", "hehe2", 11)
	dispatcher.RemoveEvent("test", t.onTest2)
	dispatcher.Empty()

	dispatcher.Dispatch("test", "hehe2", 11)

}

func (t *MClass) onTest(a string) {
	t.t.Log(a)
}

func (t *MClass) onTest2(a string, b int, c *uint) {
	t.t.Log(a, b)
}
