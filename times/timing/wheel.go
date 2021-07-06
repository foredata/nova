package timing

const (
	slotPow = 10           // 2^slotPow,默认为10
	slotMax = 1 << slotPow // 每个wheel中slot最大个数
)

func newWheel(index int) *wheel {
	w := &wheel{}
	w.init(index)
	return w
}

type wheel struct {
	slots   []*bucket // 所有桶
	index   int       // 当前索引
	offset  int       // 偏移位数
	maximum int64     // 最大取值
}

func (w *wheel) init(index int) {
	w.index = 0
	w.offset = index * slotPow
	w.maximum = int64(uint64(1) << ((index + 1) * slotPow))
	for i := 0; i < slotMax; i++ {
		w.slots = append(w.slots, newBucket())
	}
}

// 返回当前指向的桶
func (w *wheel) Current() *bucket {
	return w.slots[w.index]
}

// 向前前进一格,到最后一格则归零,并返回true
func (w *wheel) Step() bool {
	w.index++
	if w.index >= slotMax {
		w.index = 0
		return true
	}

	return false
}

// 添加到正确的桶当中
func (w *wheel) Push(t *timer, delta int64) {
	off := int(delta>>w.offset) - 1
	index := (off + w.index) % slotMax
	w.slots[index].Push(t)
}
