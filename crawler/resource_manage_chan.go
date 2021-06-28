package crawler

type ResourceManageChan struct {
	capnum uint
	mc     chan uint
}

func NewResourceManageChan(num uint) *ResourceManageChan {
	mc := make(chan uint, num)
	return &ResourceManageChan{mc: mc, capnum: num}
}

func (this *ResourceManageChan) GetOne() {
	this.mc <- 1
}

func (this *ResourceManageChan) FreeOne() {
	<-this.mc
}

func (this *ResourceManageChan) Has() uint {
	return uint(len(this.mc))
}

func (this *ResourceManageChan) Left() uint {
	return this.capnum - uint(len(this.mc))
}
