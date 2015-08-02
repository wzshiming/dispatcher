package dispatcher

type Cbs struct {
	cbs []func()
}

func NewCbs() *Cbs {
	return &Cbs{
		cbs: []func(){},
	}
}

func (cb *Cbs) Push(f func()) {
	cb.cbs = append(cb.cbs, f)
}

func (cb *Cbs) join() {
	c := cb.cbs
	cb.cbs = []func(){}
	for _, v := range c {
		v()
	}
}

func (cb *Cbs) Join() {
	for len(cb.cbs) != 0 {
		cb.join()
	}
}

type Line struct {
}

func NewLine() *Line {
	return &Line{}
}

func (*Line) Push(f func()) {
	f()
}

func (*Line) Join() {
}
