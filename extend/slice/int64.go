package cherrySlice

type Int64 []int64

func (p *Int64) Add(values ...int64) bool {
	*p = append(*p, values...)
	return true
}

func (p *Int64) Insert(index int, values ...int64) {
	if cap(*p) >= len(*p)+len(values) {
		*p = (*p)[:len(*p)+len(values)]
	} else {
		*p = append(*p, values...)
	} // else
	copy((*p)[index+len(values):], (*p)[index:])
	copy((*p)[index:], values[:])
}

func (p Int64) In(value int64) bool {
	_, result := p.IndexOf(value)
	return result
}

func (p Int64) IndexOf(value int64) (int, bool) {
	for i, vv := range p {
		if vv == value {
			return i, true
		}
	}
	return 0, false
}

func (p *Int64) RemoveIndex(index int) (int64, bool) {
	if index < 0 || index > len(*p) {
		return 0, false
	}

	v := (*p)[index]
	*p = append((*p)[0:index], (*p)[index+1:]...)

	return v, true
}

func (p *Int64) Remove(value int64) bool {
	for i, v := range *p {
		if v == value {
			_, result := p.RemoveIndex(i)
			return result
		}
	}

	return false
}

func (p Int64) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Int64) Equals(t []int64) bool {
	if len(p) != len(t) {
		return false
	}

	for i := range p {
		if p[i] != t[i] {
			return false
		}
	}

	return true
}

func (p *Int64) Clear() {
	*p = (*p)[:0]
}

func (p Int64) IsEmpty() bool {
	return len(p) < 1
}
