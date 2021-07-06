package sqlx

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

// Register makes a database driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("sql: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("sql: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]string, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a
// database name and connection information.
//
// Most users will open a database via a driver-specific connection
// helper function that returns a *DB. No database drivers are included
// in the Go standard library. See https://golang.org/s/sqldrivers for
// a list of third-party drivers.
//
// Open may just validate its arguments without creating a connection
// to the database. To verify that the data source name is valid, call
// Ping.
//
// The returned DB is safe for concurrent use by multiple goroutines
// and maintains its own pool of idle connections. Thus, the Open
// function should be called just once. It is rarely necessary to
// close a DB.
func Open(driverName, dataSourceName string, opts ...*OpenOptions) (Conn, error) {
	driversMu.RLock()
	driveri, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		// default as sql driver, check raw drivier exists
		exists := false
		allDrivers := sql.Drivers()
		for _, name := range allDrivers {
			if name == driverName {
				exists = true
				break
			}
		}

		if !exists {
			return nil, fmt.Errorf("sql: unknown driver %q (forgotten import?)", driverName)
		}

		driveri = &sqlDriver{rawDriverName: driverName}
	}

	return driveri.Open(dataSourceName, toOpenOptions(opts...))
}

// D is an ordered representation of a BSON document. This type should be used when the order of the elements matters,
// such as MongoDB command documents. If the order of the elements does not matter, an M should be used instead.
//
// Example usage:
//
// 		bson.D{{"foo", "bar"}, {"hello", "world"}, {"pi", 3.14159}}
type D []E

// Map creates a map from the elements of the D.
func (d D) Map() M {
	m := make(M, len(d))
	for _, e := range d {
		m[e.Key] = e.Value
	}
	return m
}

// E represents a BSON element for a D. It is usually used inside a D.
type E struct {
	Key   string
	Value interface{}
}

// M is an unordered representation of a BSON document. This type should be used when the order of the elements does not
// matter. This type is handled as a regular map[string]interface{} when encoding and decoding. Elements will be
// serialized in an undefined, random order. If the order of the elements matters, a D should be used instead.
//
// Example usage:
//
// 		bson.M{"foo": "bar", "hello": "world", "pi": 3.14159}.
type M map[string]interface{}

// Driver .
type Driver interface {
	Open(dataSourceName string, opts *OpenOptions) (Conn, error)
}

// Conn 代表客户端连接
type Conn interface {
	CreateDB(ctx context.Context, name string) error
	DropDB(ctx context.Context, name string) error
	Database(ctx context.Context, name string) (DB, error)
	// ListDatabases(ctx context.Context) ([]string, error)
	Ping(ctx context.Context) error
}

// DB db相关接口,filter使用interface{},可以是bson.M或者Filter
type DB interface {
	Name() string

	Close() error

	Indexes(ctx context.Context, table string) ([]*Index, error)
	CreateIndex(ctx context.Context, table string, index *Index) error
	DropIndex(ctx context.Context, table string, name string) error

	CreateTable(ctx context.Context, table string, columns []*Column) error
	DropTable(ctx context.Context, table string, opts ...*DropTableOptions) error

	Begin(ctx context.Context, opts *TxOptions) (Tx, error)

	// CRUD接口,Options仅支持1个
	Insert(ctx context.Context, table string, datas []interface{}, opts ...*InsertOptions) (Result, error)
	InsertOne(ctx context.Context, table string, data interface{}, opts ...*InsertOptions) (Result, error)
	Update(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	// Delete 若filter为nil则删除所有数据，但不删除table
	Delete(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	Find(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) Cursor
	FindOne(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) SingleResult
	// Aggregate(ctx context.Context, pipeline interface{}) (Cursor, error)

	// 原生sql
	Query(ctx context.Context, query string, args ...interface{}) Cursor
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
}

// Tx 事务
type Tx interface {
	Commit() error
	Rollback() error

	// 通用CRUD
	Insert(ctx context.Context, table string, datas []interface{}, opts ...*InsertOptions) (Result, error)
	InsertOne(ctx context.Context, table string, data interface{}, opts ...*InsertOptions) (Result, error)
	Update(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	Delete(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	Find(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) Cursor
	FindOne(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) SingleResult

	// 原生sql
	Query(ctx context.Context, query string, args ...interface{}) Cursor
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
}

type Hook interface {
}

// Filter .
type Filter interface {
	Query() string
	Args() []interface{}
}

// Cursor 多条查询结果
type Cursor interface {
	// Next 返回是否还有记录
	Next() bool
	// Decode 解析单条数据到out中,out可以是struct指针,可以是map或数组指针
	Decode(out interface{}) error
	// All 解析所有数据到out中,out为slice指针
	All(out interface{}) error
	// Close .
	Close() error
	// Error .
	Error() error
}

// SingleResult represents a single document returned from an operation. If the operation resulted in an error, all
// SingleResult methods will return that error. If the operation did not return any documents, all SingleResult methods
// will return ErrNoDocuments.
type SingleResult interface {
	Decode(out interface{}) error
	Error() error
}

type Result interface {
	InsertIDs() []interface{}
	RowsAffected() int64
}

// ReadPreference see: https://docs.mongodb.com/manual/core/read-preference/
type ReadPreference uint8

const (
	_ ReadPreference = iota
	// ReadPrefPrimary indicates that only a primary is
	// considered for reading. This is the default
	// mode.
	ReadPrefPrimary
	// ReadPrefPrimaryPreferred indicates that if a primary
	// is available, use it; otherwise, eligible
	// secondaries will be considered.
	ReadPrefPrimaryPreferred
	// ReadPrefSecondary indicates that only secondaries
	// should be considered.
	ReadPrefSecondary
	// ReadPrefSecondaryPreferred indicates that only secondaries
	// should be considered when one is available. If none
	// are available, then a primary will be considered.
	ReadPrefSecondaryPreferred
	// ReadPrefNearest indicates that all primaries and secondaries
	// will be considered.
	ReadPrefNearest
)

type OrderType int

const (
	OrderAsc  = 0 // 默认升序
	OrderDesc = 1
)

func (o OrderType) String() string {
	if o == OrderAsc {
		return "ASC"
	} else {
		return "DESC"
	}
}

type Order struct {
	Field string
	Order OrderType
}

type IndexKey struct {
	Name  string
	Order OrderType
}

// Index 索引信息
type Index struct {
	Name       string     // 索引名,若为空则使用Keys拼接
	Keys       []IndexKey // 索引key
	Background bool       // 是否后台异步创建索引
	Unique     bool       // 是否唯一索引
	Sparse     bool       // 稀疏索引,see mongno: https://mongoing.com/docs/core/index-sparse.html
}

func (idx *Index) toName() string {
	if idx.Name != "" {
		return idx.Name
	}

	b := bytes.NewBuffer(nil)
	for _, k := range idx.Keys {
		if k.Name == "" {
			panic(fmt.Errorf("sqlx: invalid index key name"))
		}
		if b.Len() > 0 {
			b.WriteByte('_')
		}
		b.WriteString(k.Name)
		b.WriteByte('_')
		if k.Order == OrderAsc {
			b.WriteByte('0')
		} else {
			b.WriteByte('1')
		}
	}

	idx.Name = b.String()

	return idx.Name
}

type FieldType int

// 支持的数据类型
const (
	FTChar FieldType = iota
	FTInt8
	FTInt16
	FTInt32
	FTInt64
	FTUint8
	FTUint16
)

// Column 列信息,用于CreateTable
type Column struct {
	Name          string // 字段名
	Type          string // 类型
	Default       string // 默认值, 不需要双引号
	NotNull       bool   // 是否非空
	AutoIncrement bool   // 是否自增
	PrimaryKey    bool   // 是否是主键
	Size          int    // 类型大小
}
