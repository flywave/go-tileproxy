package crawler

type ResourceManage interface {
	GetOne()
	FreeOne()
	Has() uint
	Left() uint
}
