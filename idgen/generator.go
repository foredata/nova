package idgen

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	// ErrTimeBack 时间发生回拨
	ErrTimeBack = errors.New("idgen: time go back")
	// ErrTooBigNodeID node超出范围,10bits
	ErrTooBigNodeID = errors.New("idgen: too big note id")
	// ErrSequenceOverflow SeqID溢出,等待下一个时间段重新生成
	ErrSequenceOverflow = errors.New("idgen: sequence overflow")
	// ErrTimestampOverflow 时间戳溢出
	ErrTimestampOverflow = errors.New("idgen: timestamp overflow")
	// ErrInvalidBase32 无效的base32
	ErrInvalidBase32 = errors.New("idgen: invalid base32")
	// ErrInvalidBase58 无效的base58
	ErrInvalidBase58 = errors.New("idgen: invalid base58")
)

// startTime 起始时间
var startTime = time.Date(2020, time.January, 0, 0, 0, 0, 0, time.UTC)

// Generator id生成器接口
type Generator interface {
	Next() (ID, error)
}

// NewGenerator 创建Generator
// 在分布式环境下,需要nodeID唯一,这里提供几种方案
//	1:redis中记录一共值,每次递增取模%MaxNodeID
//	2:基于分布式锁,服务启动时从[0-MaxNodeID)轮训取ID并注册,比如redis,zookeeper,etcd,consul
//	3:手动指定,通过IP等静态映射
func NewGenerator(nodeID uint, opts ...Option) Generator {
	g := &generator{}
	err := g.Init(nodeID, newOptions(opts...))
	if err != nil {
		panic(err)
	}
	return g
}

type ID int64

func toID(isSec bool, ts uint64, seqId uint64, nodeId uint64) ID {
	if isSec {
		return ID(1<<62 | ts<<30 | seqId<<10 | nodeId)
	} else {
		return ID(ts<<22 | seqId<<10 | nodeId)
	}
}

func (id ID) IsSecond() bool {
	return (id & (1 << 61)) != 0
}

func (id ID) Time() time.Time {
	if id.IsSecond() {
		sec := int64((id>>30)&^(1<<32)) + startTime.Unix()
		return time.Unix(sec, 0)
	} else {
		ms := int64(id>>22) + startTime.UnixNano()/int64(time.Millisecond)
		return time.Unix(ms/1000, (ms%1000)*1000000)
	}
}

func (id ID) Sequence() uint64 {
	if id.IsSecond() {
		return uint64(id>>10) & (1<<20 - 1)
	} else {
		return uint64(id>>10) & (1<<12 - 1)
	}
}

func (id ID) NodeID() uint64 {
	return uint64(id) & (1<<10 - 1)
}

func (id ID) Zero() bool {
	return id == 0
}

func (id ID) Hex() string {
	return strconv.FormatUint(uint64(id), 16)
}

func (id ID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id ID) Bytes() []byte {
	return []byte(id.String())
}

func (id ID) Base32() string {
	return toBase32(uint64(id))
}

func (id ID) Base58() string {
	return toBase58(uint64(id))
}

func (id ID) Base64() string {
	return toBase64(id.Bytes())
}

// MarshalJSON returns a json byte array string of the ID.
func (id ID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(id), 10)
	buff = append(buff, '"')
	return buff, nil
}

// UnmarshalJSON converts a json byte array of a snowflake ID into an ID type.
func (id *ID) UnmarshalJSON(b []byte) error {
	if len(b) < 3 || b[0] != '"' || b[len(b)-1] != '"' {
		return fmt.Errorf("invalid id %q", string(b))
	}

	i, err := strconv.ParseUint(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}

	*id = ID(i)
	return nil
}

// 以数字形式展示,年月日-时分秒-毫秒-序列号-机器ID-类型
func (id ID) Format() string {
	ts := id.Time()
	seq := id.Sequence()
	nodeID := id.NodeID()
	if id.IsSecond() {
		return fmt.Sprintf("%04d%02d%02d-%02d%02d%02d-%d-%d-%d",
			ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(),
			seq, nodeID, 1)
	} else {
		return fmt.Sprintf("%04d%02d%02d-%02d%02d%02d-%04d-%d-%d-%d",
			ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Nanosecond()/1000000,
			seq, nodeID, 0)
	}
}

type Clocker interface {
	Now() time.Time
}

const (
	// MaxNodeID 最大noteId
	MaxNodeID = 1 << 10
)

// generator 默认id生成器,类似snowflake,支持秒和毫秒两种模式
// 秒级:| 1 Bit Unused | 1 Bit TimeMode(1) | 32 Bit Timestamp | 20 Bit SeqID | 10 Bit NodeID
// 毫秒:| 1 Bit Unused | 1 Bit TimeMode(0) | 40 Bit Timestamp | 12 Bit SeqID | 10 Bit NodeID
// 注:
// 由于强依赖时钟,时钟回调会导致ID重复,在毫秒模式下会更严重,而秒模式下因为回调时间比较短,可能并不明显,
// ID生成器需要关闭NTP同步,如果发生时间回调,生成ID会报错
type generator struct {
	lock         sync.Locker //
	clock        Clocker     //
	template     uint64      // 生成模板,由noteid+time mode
	precision    int64       // 时间精度,秒/毫秒
	timestamp    uint64      // 当前时间戳
	timestampOff uint64      // 时间戳偏移
	timestampMax uint64      // 时间戳最大值
	sequence     uint64      // 当前序列号
	sequenceOff  uint64      // sequeue偏移
	sequenceMax  uint64      // sequeue最大值
}

func (g *generator) Init(nodeId uint, opts *Options) error {
	if nodeId > MaxNodeID {
		return ErrTooBigNodeID
	}

	g.lock = opts.lock
	g.clock = opts.clock

	if opts.useSecond {
		// 秒模式
		g.precision = int64(time.Second)
		g.sequenceMax = 1 << 20
		g.timestampMax = 1 << 32
		g.timestampOff = 30
		g.sequenceOff = 10
		g.template = 1<<62 + uint64(nodeId)
	} else {
		g.precision = int64(time.Millisecond)
		g.sequenceMax = 1 << 12
		g.timestampMax = 1 << 40
		g.timestampOff = 22
		g.sequenceOff = 10
		g.template = uint64(nodeId)
	}

	return nil
}

func (g *generator) Next() (ID, error) {
	g.lock.Lock()
	defer g.lock.Unlock()
	ts := uint64(g.clock.Now().Sub(startTime).Nanoseconds() / int64(g.precision))
	if ts >= g.timestampMax {
		return 0, ErrTimestampOverflow
	}

	switch {
	case ts < g.timestamp:
		return 0, ErrTimeBack
	case ts > g.timestamp:
		g.timestamp = ts
		g.sequence = 0
	default:
		g.sequence++
		if g.sequence >= g.sequenceMax {
			return 0, ErrSequenceOverflow
		}
	}

	return ID(g.template | ts<<g.timestampOff | g.sequence), nil
}
