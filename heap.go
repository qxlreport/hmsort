package hmsort

type partRec struct {
	buf    []byte
	offset int64
}

type partHeap struct {
	ph []*partRec
	ls Less
}

func (pr partHeap) Len() int {
	return len(pr.ph)
}

func (pr partHeap) Less(i, j int) bool {
	return pr.ls(pr.ph[i].buf, pr.ph[j].buf)
}

func (pr partHeap) Swap(i, j int) {
	pr.ph[i], pr.ph[j] = pr.ph[j], pr.ph[i]
}

func (pr *partHeap) Push(x interface{}) {
	(*pr).ph = append((*pr).ph, x.(*partRec))
}

func (pr *partHeap) Pop() interface{} {
	old := (*pr).ph
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	(*pr).ph = old[0 : n-1]
	return item
}
