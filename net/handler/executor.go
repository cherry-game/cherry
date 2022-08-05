package cherryHandler

type Executor struct {
	groupIndex int
}

func (p *Executor) SetIndex(index int) {
	p.groupIndex = index
}

func (p *Executor) Index() int {
	return p.groupIndex
}
