package bytex

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

var (
	ErrNotSupport   = errors.New("[buffer] not support")
	ErrInvalidParam = errors.New("[buffer] invalid param")
)

var (
	chunkSize = 1024 // 默认分配内存大小
)

const (
	// NPOS 无效索引
	NPOS = -1
)

// SetChunkSize 修改默认分配大小,大小应是2的幂,大小标准应尽量与消息包大小保持一致
func SetChunkSize(size int) {
	chunkSize = size
}

// Peeker 类似io.Reader但不会修改游标
type Peeker interface {
	Peek(data []byte) (int, error)
}

// BytePeeker 类似io.ByteReader,但不会修改游标,无数据会返回io.EOF错误
type BytePeeker interface {
	PeekByte() (byte, error)
}

// ReaderFromOnce 从io.Reader中读取一次数据
//	区别于io.ReaderFrom,这里不管成功或失败,只会读取一次,
//	之所以这样是因为io.Reader可能是阻塞的,比如net.Conn
type ReaderFromOnce interface {
	ReadFromOnce(r io.Reader) (n int64, err error)
}

// Buffer 是一个比较底层的缓冲区接口，主要用于网络数据交换
//	特点:
//	1:基于链表,非连续,避免多次内存拷贝
//	2:基于内存池,减少gc(效果待验证),在Buffer使用完成后,需要主动调用Clear,才能回收复用内存
//	3:基于引用计数,便于安全的共享内存,但不支持COW(copy on write),多个buffer写入相同内存会被覆盖
//	4:单游标操作,区别于netty的ByteBuf的读写分离方式,这里只有一个position,使用时需要seek到正确位置
//	比如:一次消息的处理过程：
//	SeekToEnd-> 接受数据 --> SeekToStart --> 读取数据并解码处理-->循环这个过程
//	5:特殊函数
//	Append和Prepend并不基于当前游标读写,而是直接在尾部和头部修改
//
// 区别于Netty中的ByteBuf,不区分readIndex和writeIndex,只有当前位置
// https://netty.io/4.1/api/index.html
//
// +-------------+--------------+---------------+
// | 		data(read/write)	|  unused bytes |
// +-------------+--------------+---------------+
// |             |      		|  		   	    |
// 0      <=    pos	  <=	 length    <=  capacity
type Buffer interface {
	Peeker
	BytePeeker
	io.Closer
	io.Seeker
	io.Writer
	io.Reader
	io.ByteReader
	io.ByteWriter
	io.WriterTo
	ReaderFromOnce
	Len() int                       // 数据长度
	Cap() int                       // 真实容量
	Pos() int                       // 当前位置
	Available() int                 // 可用数据大小[len-pos]
	Empty() bool                    // 是否为空[len==0]
	Bytes() []byte                  // 转为[]byte
	String() string                 // 转为字符串
	Clear()                         // 清空内存并回收复用
	Discard()                       // 丢弃当前位置之前的数据
	Append(data interface{}) error  // 在尾部追加数据,并将pos移动到末尾
	Prepend(data interface{}) error // 在头部插入数据,并将pos移动开始
	WriteN(n int)                   // 从当前位置写入n个字节,移动游标,但不会写入数据
	PeekN(n int) Buffer             // 从当前位置获取n个字符,不修改游标,数据不足返回nil
	ReadN(n int) Buffer             // 从当前位置读取n个字符,数据不足返回nil
	ReadLine() Buffer               // 从当前位置读取一行(\n|\r\n), 不存在返回nil
	IndexByte(c byte, max int) int  // 从当前位置查找byte,max<=0则查询全部,返回相对当前位置的偏移,不存在返回-1
}

// NewBuffer 创建Buffer,底层使用连表,非连续
func NewBuffer() Buffer {
	return newBuffer()
}

var gNodePool = sync.Pool{
	New: func() interface{} {
		return &bnode{}
	},
}

var gChunkPool = sync.Pool{
	New: func() interface{} {
		return &bchunk{}
	},
}

func newBuffer() *buffer {
	return &buffer{}
}

type buffer struct {
	head   *bnode  // 链表头
	tail   *bnode  // 链表尾
	cursor bcursor // 当前游标
	pos    int     // 当前位置
	len    int     // 总长度
	cap    int     // 总容量
}

// bchunk 基于引用计数的byte,主动释放时放到内存池中
type bchunk struct {
	data []byte //
	refs int32  //
}

func newChunk(data []byte) *bchunk {
	chunk := gChunkPool.Get().(*bchunk)
	chunk.data = data
	return chunk
}

func (b *bchunk) Obtain() {
	atomic.AddInt32(&b.refs, 1)
}

func (b *bchunk) Free() {
	if atomic.AddInt32(&b.refs, -1) == 0 {
		Free(b.data)
		gChunkPool.Put(b)
	}
}

// bnode 用于实现双向非循环链表
type bnode struct {
	prev *bnode
	next *bnode
	data []byte  // 当前数据
	link *bchunk // 原始数据,free时会回收
}

// newNode 通过原始数据创建node
func newNode(data []byte, chunk *bchunk) *bnode {
	if chunk != nil {
		chunk.Obtain()
	}

	n := gNodePool.Get().(*bnode)
	n.prev = nil
	n.next = nil
	n.data = data
	n.link = chunk

	return n
}

// newNodeBySize 通过字节大小创建node
func newNodeBySize(size int) *bnode {
	data := Alloc(size)
	chunk := newChunk(data)
	return newNode(data, chunk)
}

func (b *bnode) Len() int {
	return len(b.data)
}

func (b *bnode) Back() byte {
	return b.data[len(b.data)-1]
}

func (b *bnode) Free() {
	if b.link != nil {
		b.link.Free()
		b.link = nil
	}
	gNodePool.Put(b)
}

// Clone 拷贝部分数据
func (b *bnode) Clone(start, end int) *bnode {
	data := b.data[start:end]
	return newNode(data, b.link)
}

// bcursor 游标,用于标记当前位置
type bcursor struct {
	node *bnode // 当前节点指针
	offs int    // 当前节点偏移位置
}

func (b *bcursor) Data() []byte {
	return b.node.data
}

// Byte 当前位置对应的字节
func (b *bcursor) Byte() byte {
	return b.node.data[b.offs]
}

// SetByte 当前位置设置byte
func (b *bcursor) SetByte(ch byte) {
	b.node.data[b.offs] = ch
}

// MoveByte 移动1个字节
func (b *bcursor) MoveByte() {
	if b.node.Len() == b.offs {
		b.node = b.node.next
		b.offs = 0
	} else {
		b.offs++
	}
}

// reset 重置游标
func (b *bcursor) Reset(n *bnode, offs int) {
	b.node = n
	b.offs = offs
}

// Skip 如果index到达结尾,则自动跳转到下一个节点
func (b *bcursor) Skip() {
	if b.offs == len(b.node.data) && b.node.next != nil {
		b.node = b.node.next
		b.offs = 0
	}
}

/////////////////////////////////////////////
// buffer
/////////////////////////////////////////////
func (b *buffer) Empty() bool {
	return b.len == 0
}

func (b *buffer) Len() int {
	return b.len
}

func (b *buffer) Cap() int {
	return b.cap
}

func (b *buffer) Pos() int {
	return b.pos
}

// Available 有效数据大小
func (b *buffer) Available() int {
	return b.len - b.pos
}

func (b *buffer) Bytes() []byte {
	if b.len > 0 {
		b.concat()
		return b.head.data[:b.len]
	}

	return nil
}

func (b *buffer) String() string {
	if b.len > 0 {
		b.concat()
		return string(b.head.data[:b.len])
	}

	return ""
}

// concat 合并成一块内存
func (b *buffer) concat() {
	if b.head == b.tail {
		return
	}

	n := newNodeBySize(b.len)
	copyTo(bcursor{b.head, 0}, b.len, n.data)
	b.freeNodes()
	b.head = n
	b.tail = n
	b.cursor.Reset(n, b.pos)
}

////////////////////////////////////////////////////
// io相关接口
////////////////////////////////////////////////////
// Close 实现io.Closer接口
func (b *buffer) Close() error {
	b.Clear()
	return nil
}

// see: io.Seeker
func (b *buffer) Seek(offset int64, whence int) (int64, error) {
	// 计算当前位置
	var pos int
	switch whence {
	case io.SeekCurrent:
		// offset为负数表示反向移动游标
		pos = b.pos + int(offset)
	case io.SeekStart:
		if offset < 0 {
			return 0, fmt.Errorf("SeekStart offset < 0")
		}
		pos = int(offset)
	case io.SeekEnd:
		if offset < 0 {
			return 0, fmt.Errorf("SeekEnd offset < 0")
		}
		pos = b.len - int(offset)
	}

	if pos < 0 || pos > b.len {
		return 0, ErrInvalidParam
	}

	if b.cursor.node == nil {
		whence = io.SeekStart
	} else if pos == b.pos {
		return int64(b.pos), nil
	}

	// fast seek
	switch {
	case pos == 0:
		b.cursor.Reset(b.head, 0)
		b.pos = 0
		return int64(b.pos), nil
	case pos == b.len:
		offset := b.tail.Len() - (b.cap - b.len)
		b.cursor.Reset(b.tail, offset)
		b.pos = b.len
		return int64(b.pos), nil
	}

	// slow seek
	switch whence {
	case io.SeekCurrent:
		if offset > 0 { // 向后移动
			b.cursor = moveNextTo(b.cursor, int(offset))
		} else { // 向前移动
			b.cursor = movePrevTo(b.cursor, int(-offset))
		}
	case io.SeekStart:
		b.cursor = moveNextTo(bcursor{b.head, 0}, pos)
	case io.SeekEnd:
		b.cursor = movePrevTo(bcursor{b.tail, len(b.tail.data)}, b.cap-pos)
	}

	b.pos = pos
	return int64(pos), nil
}

// Peek 类似于Read,但是不会修改当前游标
func (b *buffer) Peek(p []byte) (int, error) {
	size := len(p)
	if size == 0 || size > b.Available() {
		return 0, io.EOF
	}

	b.ensureCursor()
	copyTo(b.cursor, size, p)

	return size, nil
}

// Read 读取数据
//	1: len(p) == 0, 返回0, nil
//	2: len(p) <  usable: 返回len(p), nil
//	3: len(p) == usable: 返回len(p), nil,下次再读取返回0,io.EOF
//	4: len(p) >  usable: 返回usable, io.EOF
func (b *buffer) Read(p []byte) (n int, err error) {
	size := len(p)
	usable := b.Available()
	switch {
	case size == 0:
		return 0, nil
	case size <= usable:
		n = size
		err = nil
	default:
		n = usable
		err = io.EOF
	}

	b.ensureCursor()
	b.cursor = copyTo(b.cursor, n, p)
	b.pos += n
	return
}

// Write 从当前位置写入数据,并向后移动游标,若空间不足会先分配空间
func (b *buffer) Write(p []byte) (int, error) {
	size := len(p)
	if size == 0 {
		return 0, nil
	}

	b.ensureSize(size)
	b.ensureCursor()

	b.cursor = copyFrom(b.cursor, size, p)
	b.pos += size

	return size, nil
}

// PeekByte 从当前位置读取一个字节,不修改游标
func (b *buffer) PeekByte() (byte, error) {
	if b.Available() == 0 {
		return 0, io.EOF
	}
	b.ensureCursor()
	ch := b.cursor.Byte()
	return ch, nil
}

// ReadByte 读取一个字节,并修改游标,无数据返回io.EOF
func (b *buffer) ReadByte() (byte, error) {
	if b.Available() == 0 {
		return 0, io.EOF
	}

	b.ensureCursor()
	ch := b.cursor.Byte()
	b.cursor.MoveByte()
	b.pos++
	return ch, nil
}

// WriteByte 写入一个字节
func (b *buffer) WriteByte(ch byte) error {
	b.ensureSize(1)
	b.ensureCursor()
	b.cursor.SetByte(ch)
	b.cursor.MoveByte()
	b.pos++
	return nil
}

// WriteTo 实现io.WriterTo接口,从当前位置将数据写入target,并修改游标
func (b *buffer) WriteTo(w io.Writer) (int64, error) {
	if b.len-b.pos == 0 {
		return 0, nil
	}

	var n int
	var err error
	iter := newIterator(b.cursor, b.len-b.pos)
	for iter.Next() {
		n, err = w.Write(iter.data)
		if n < len(iter.data) {
			iter.Rollback(n)
			break
		}
	}

	b.cursor = iter.bcursor
	b.pos += iter.curr
	return int64(iter.curr), err
}

// ReadFromOnce 从数据源读取数据,并从当前位置写入,同时修改游标
func (b *buffer) ReadFromOnce(r io.Reader) (int64, error) {
	if b.pos == b.cap {
		b.grow(chunkSize)
	}
	b.ensureCursor()

	data := b.cursor.Data()
	offs := b.cursor.offs
	n, err := r.Read(data[offs:])
	b.pos += n
	b.cursor.offs += n

	if b.pos > b.len {
		b.len = b.pos
	}

	return int64(n), err
}

func (b *buffer) Clear() {
	b.freeNodes()
	b.cursor.Reset(nil, 0)
	b.pos = 0
	b.len = 0
	b.cap = 0
}

// Discard 丢弃pos之前的数据,并将当前位置设置为0
func (b *buffer) Discard() {
	if b.pos == 0 {
		return
	}

	if b.pos == b.len {
		b.Clear()
		return
	}

	count := b.pos
	for count > 0 {
		node := b.head
		size := len(node.data)
		if size > count {
			node.data = node.data[count:]
			break
		}

		count -= size
		b.head = node.next
		node.Free()
	}

	b.cap -= b.pos
	b.len -= b.pos
	b.pos = 0
	b.cursor.Reset(b.head, 0)
}

func (b *buffer) WriteN(n int) {
	if b.cap-b.pos < n {
		b.grow(n)
	}
	b.pos += n
	if b.pos > b.len {
		b.len = b.pos
	}
	if b.cursor.node != nil {
		b.cursor = moveNextTo(b.cursor, n)
	}
}

func (b *buffer) Prepend(data interface{}) error {
	switch x := data.(type) {
	case *buffer:
		if x.Len() > 0 {
			for n := x.tail; n != nil; n = n.prev {
				b.pushFront(n.data, n.link)
			}
		}
	case []byte:
		if len(x) > 0 {
			b.pushFront(x, nil)
		}
	case string:
		if len(x) > 0 {
			b.pushFront([]byte(x), nil)
		}
	}

	return nil
}

// Append 末尾追加数据,并将pos移动到末尾
func (b *buffer) Append(data interface{}) error {
	b.freeUnusedBytes()
	switch x := data.(type) {
	case *buffer:
		if x.Len() > 0 {
			size := x.len
			for n := x.head; n != nil; n = n.next {
				if len(n.data) < size {
					b.pushBack(n.data, n.link, true)
					size -= len(n.data)
				} else {
					b.pushBack(n.data[:size], n.link, true)
					break
				}
			}
		}
	case []byte:
		if len(x) > 0 {
			b.pushBack(x, nil, true)
		}
	case string:
		if len(x) > 0 {
			b.pushBack([]byte(x), nil, true)
		}
	default:
		return ErrNotSupport
	}
	// move to end
	b.pos = b.len
	b.cursor.Reset(nil, 0)
	return nil
}

func (b *buffer) PeekN(n int) Buffer {
	if n == 0 || n > b.Available() {
		return nil
	}

	b.ensureCursor()

	result := newBuffer()
	iter := newIterator(b.cursor, n)
	for iter.Next() {
		result.pushBack(iter.data, iter.node.link, true)
	}
	return nil
}

// ReadN 读取n个字节,不足则返回nil
func (b *buffer) ReadN(n int) Buffer {
	if n == 0 || n > b.Available() {
		return nil
	}

	b.ensureCursor()

	result := newBuffer()
	iter := newIterator(b.cursor, n)
	for iter.Next() {
		result.pushBack(iter.data, iter.node.link, true)
	}
	b.cursor = iter.bcursor
	b.pos += n
	return result
}

// ReadLine 读取一行数据\n或\r\n,返回结果不包含分隔符
func (b *buffer) ReadLine() Buffer {
	b.ensureCursor()
	result := newBuffer()
	iter := newIterator(b.cursor, b.len-b.pos)
	for iter.Next() {
		index := iter.IndexByte('\n')
		if index != -1 {
			if index > 0 {
				result.pushBack(iter.data[:index], iter.node.link, true)
			}
			// 去除\r
			if result.tail != nil && result.tail.Back() == '\r' {
				result.len--
				if result.tail.Len() == 1 {
					result.popBack()
				}
			}
			iter.Rollback(index + 1)
			b.cursor = iter.bcursor
			b.pos += iter.curr
			return result
		}
		result.pushBack(iter.data, iter.node.link, true)
	}

	// 不存在
	result.Clear()
	return nil
}

func (b *buffer) IndexByte(c byte, max int) int {
	if max <= 0 {
		max = b.len - b.pos
	}

	iter := newIterator(b.cursor, max)
	for iter.Next() {
		index := iter.IndexByte(c)
		if index != -1 {
			return iter.last + index
		}
	}

	return -1
}

// freeUnusedBytes 清空末尾无效的空间
func (b *buffer) freeUnusedBytes() {
	if b.tail == nil {
		return
	}
	n := b.cap - b.len
	if n == 0 {
		return
	}

	if b.len == 0 {
		b.Clear()
		return
	}

	node := b.tail
	for n > 0 {
		x := len(node.data)
		if x > n {
			node.data = node.data[:x-n]
			b.tail = node
			break
		}

		n -= x
		t := node
		node = node.prev
		t.Free()
	}
}

// grow 扩容size个字节
func (b *buffer) grow(size int) {
	if size < chunkSize {
		size = chunkSize
	}
	data := Alloc(size)
	chunk := newChunk(data)
	b.pushBack(data, chunk, false)
}

// ensure 确保相对当前位置,有足够空间可以写入数据
func (b *buffer) ensureSize(size int) {
	end := b.pos + size
	if end <= b.len {
		return
	}

	if end > b.cap {
		b.grow(end - b.cap)
	}
	b.len = end
}

// ensureCursor 确保当前游标不为nil,并自动调过到达末尾的游标
func (b *buffer) ensureCursor() {
	if b.cursor.node == nil {
		_, _ = b.Seek(int64(b.pos), io.SeekStart)
	}
	b.cursor.Skip()
}

// pushBack 在链表尾部插入节点,会自动扩展cap
// 有两种使用场景
//	1:扩展容量,此时只会增加cap
//	2:追加数据,此时需要保证len==cap,会同时增加cap和len
func (b *buffer) pushBack(data []byte, chunk *bchunk, addLen bool) {
	if len(data) == 0 {
		return
	}
	n := newNode(data, chunk)
	if b.tail != nil {
		t := b.tail
		t.next = n
		n.prev = t
		b.tail = n
	} else {
		b.head = n
		b.tail = n
	}
	if addLen {
		b.len += len(data)
	}
	b.cap += len(data)
}

// popBack 释放最后一个
func (b *buffer) popBack() {
	tail := b.tail
	if b.tail == b.head {
		b.head = nil
		b.tail = nil
	} else {
		b.tail = tail.prev
	}

	size := len(tail.data)
	sub := size - (b.cap - b.len)
	if sub > 0 {
		b.len -= sub
	}
	b.cap -= size
	tail.Free()
}

// pushFront 在链表头部插入节点
func (b *buffer) pushFront(data []byte, chunk *bchunk) {
	size := len(data)
	if size == 0 {
		return
	}
	n := newNode(data, chunk)
	if b.head != nil {
		h := b.head
		h.prev = n
		n.next = h
		b.head = n
	} else {
		b.head = n
		b.tail = n
	}
	b.len += size
	b.cap += size
	// 需要确保游标位置正确
	if b.cursor.node != nil {
		b.pos += size
	}
}

// freeNodes 释放所有节点
func (b *buffer) freeNodes() {
	for n := b.head; n != nil; {
		t := n
		n = n.next
		t.Free()
	}

	b.head = nil
	b.tail = nil
}

///////////////////////////////////////////
// Iterator
///////////////////////////////////////////
// copyFrom 从外部数据拷贝到内部
func copyFrom(cursor bcursor, size int, p []byte) bcursor {
	iter := newIterator(cursor, size)
	for iter.Next() {
		copy(iter.data, p[iter.last:])
	}

	return iter.bcursor
}

// copyTo 从内部数据拷贝到外部
func copyTo(cursor bcursor, size int, p []byte) bcursor {
	iter := newIterator(cursor, size)
	for iter.Next() {
		copy(p[iter.last:], iter.data)
	}

	return iter.bcursor
}

// moveNextTo 向后移动到指定位置
func moveNextTo(cursor bcursor, size int) bcursor {
	iter := newIterator(cursor, size)
	for iter.Next() {
	}
	return iter.bcursor
}

// movePrevTo 向前移动到指定位置
func movePrevTo(cursor bcursor, size int) bcursor {
	iter := newIterator(cursor, size)
	for iter.Prev() {
	}
	return iter.bcursor
}

// newIterator 创建迭代器
func newIterator(cursor bcursor, size int) biterator {
	iter := biterator{bcursor: cursor, size: size, curr: 0, last: 0}
	return iter
}

// buffer 迭代器,可前向(Prev)或后向迭代(Next),会自动忽略一次边界情况
// 要求1:node中不存在长度为0的节点
// 要求2:外部保证size的合法性,即有足够的数据可以读写
type biterator struct {
	bcursor        // 游标
	size    int    // 需要处理的长度
	curr    int    // 当前处理的位置
	last    int    // 上次处理的位置
	data    []byte // 本次可以处理的数据
}

func (iter *biterator) Next() bool {
	if iter.node == nil || iter.curr >= iter.size {
		return false
	}

	iter.Skip()

	iter.last = iter.curr
	need := iter.size - iter.curr
	data := iter.Data()
	count := len(data) - iter.offs
	if count >= need {
		// 数据足够
		iter.data = data[iter.offs : iter.offs+need]
		iter.offs += need
		iter.curr += need
	} else {
		// 数据不足,需要继续
		iter.data = data[iter.offs:]
		iter.node = iter.node.next
		iter.offs = 0
		iter.curr += count
	}

	return true
}

// Prev 向前移动游标,查找可以利用的下一个节点
func (iter *biterator) Prev() bool {
	if iter.node == nil || iter.curr >= iter.size {
		return false
	}

	if iter.offs == 0 {
		// offs为0,需要移动一次,处理前一个节点
		iter.node = iter.node.prev
		if iter.node == nil {
			return false
		}
		iter.offs = len(iter.node.data)
	}

	iter.last = iter.curr
	need := iter.size - iter.curr
	data := iter.node.data
	if iter.offs >= need {
		iter.data = data[iter.offs-need : iter.offs]
		iter.curr += need
	} else {
		iter.data = data[:iter.offs]
		iter.offs = 0
		iter.curr += iter.offs
	}

	return true
}

// IndexByte 查找某个字符
func (iter *biterator) IndexByte(ch byte) int {
	for i, v := range iter.data {
		if ch == v {
			return i
		}
	}
	return -1
}

// Rollback 回退到指定索引,用于data只处理了一半情况,
//	注:并没有同步修改iter.data,因为调用完此函数后就不再使用data了
func (iter *biterator) Rollback(index int) {
	remain := len(iter.data) - index
	if remain > 0 {
		iter.offs -= remain
		iter.curr -= remain
	}
}
