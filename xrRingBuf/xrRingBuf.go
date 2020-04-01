//!!!环形缓冲的大小必须是2的幂
//单线程写,单线程读 无锁
package xrRingBuf

import "errors"

type RingBuf struct {
	size uint32 //缓冲区空间的大小，在初化时，将它向上圆整成2的幂
	in   uint32 //指向buffer中队头
	out  uint32 //指向buffer中的队尾
	buf  []byte //用于存放数据的缓存
}

//len:2的幂
//return:err
func (p *RingBuf) Create(len uint32) (err error) {
	if !isPowerOf2(len) {
		return errors.New("不是2的次幂")
	}
	p.buf = make([]byte, len)
	p.size = len
	return nil
}

//return: 返回放入的长度
//放不下的丢掉
func (p *RingBuf) Write(data []byte, len uint32) (write_len uint32) {
	//可写入长度
	write_len = min(p.size-(p.in-p.out), len)

	l := min(write_len, p.size-p.ringBufMod(p.in))

	copy(p.buf[p.ringBufMod(p.in):], data[:l])
	copy(p.buf, data[l:write_len])

	p.in += write_len
	return write_len
}

//读出数据
//return:返回取出的长度
func (p *RingBuf) Read(data []byte, len uint32) (read_len uint32) {
	read_len = min(p.in-p.out, len)

	l := min(read_len, p.size-p.ringBufMod(p.out))
	copy(data, p.buf[p.ringBufMod(p.out):p.ringBufMod(p.out)+l])
	copy(data[l:], p.buf[:read_len-l])
	p.out += read_len
	return read_len
}

/*
//查看数据,不取出
//return:返回查看的长度
func (p *RingBuf) See(data []byte, len uint32) (read_len uint32) {
	read_len = min(p.in-p.out, len)

	l := min(read_len, p.size-p.ringBufMod(p.out))
	copy(data, p.buf[p.ringBufMod(p.out):p.ringBufMod(p.out)+l])
	copy(data[l:], p.buf[:read_len-l])
	//p.out += read_len
	return read_len
}
*/
//取模运算
func (p *RingBuf) ringBufMod(n uint32) (idx uint32) {
	return n & (p.size - 1)
}

//判断一个数是否为2的次幂
func isPowerOf2(n uint32) bool {
	return (n != 0 && ((n & (n - 1)) == 0))
}

func min(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}
