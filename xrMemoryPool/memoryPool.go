package xrMemoryPool

//每次分配出去的内存单元（称为 unit 或者 cell）的大小为程序预先定义的值。
//释放内存块时，则只需要简单地挂回内存池链表中即可。又称为 “固定尺寸缓冲池”。

//固定大小,动态数量
type MemoryPool struct {
}

//len:每个大小
//cnt:数量
func (p *MemoryPool) Init(len uint32, cnt uint32) (err error) {

	return nil
}

func (p *MemoryPool) Malloc() (buf []byte, err error) {
	return buf, err
}

func (p *MemoryPool) Free(buf []byte) (err error) {
	return nil
}
