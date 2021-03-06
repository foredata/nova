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

// Conn ?????????????????????
type Conn interface {
	CreateDB(ctx context.Context, name string) error
	DropDB(ctx context.Context, name string) error
	Database(ctx context.Context, name string) (DB, error)
	// ListDatabases(ctx context.Context) ([]string, error)
	Ping(ctx context.Context) error
}

// DB db????????????,filter??????interface{},?????????bson.M??????Filter
type DB interface {
	Name() string

	Close() error

	Indexes(ctx context.Context, table string) ([]*Index, error)
	CreateIndex(ctx context.Context, table string, index *Index) error
	DropIndex(ctx context.Context, table string, name string) error

	CreateTable(ctx context.Context, table string, columns []*Column) error
	DropTable(ctx context.Context, table string, opts ...*DropTableOptions) error

	Begin(ctx context.Context, opts *TxOptions) (Tx, error)

	// CRUD??????,Options?????????1???
	Insert(ctx context.Context, table string, datas []interface{}, opts ...*InsertOptions) (Result, error)
	InsertOne(ctx context.Context, table string, data interface{}, opts ...*InsertOptions) (Result, error)
	Update(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	// Delete ???filter???nil????????????????????????????????????table
	Delete(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	Find(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) Cursor
	FindOne(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) SingleResult
	// Aggregate(ctx context.Context, pipeline interface{}) (Cursor, error)

	// ??????sql
	Query(ctx context.Context, query string, args ...interface{}) Cursor
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
}

// Tx ??????
type Tx interface {
	Commit() error
	Rollback() error

	// ??????CRUD
	Insert(ctx context.Context, table string, datas []interface{}, opts ...*InsertOptions) (Result, error)
	InsertOne(ctx context.Context, table string, data interface{}, opts ...*InsertOptions) (Result, error)
	Update(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error)
	Delete(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error)
	Find(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) Cursor
	FindOne(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) SingleResult

	// ??????sql
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

// Cursor ??????????????????
type Cursor interface {
	// Next ????????????????????????
	Next() bool
	// Decode ?????????????????????out???,out?????????struct??????,?????????map???????????????
	Decode(out interface{}) error
	// All ?????????????????????out???,out???slice??????
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
	OrderAsc  = 0 // ????????????
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

// Index ????????????
type Index struct {
	Name       string     // ?????????,??????????????????Keys??????
	Keys       []IndexKey // ??????key
	Background bool       // ??????????????????????????????
	Unique     bool       // ??????????????????
	Sparse     bool       // ????????????,see mongno: https://mongoing.com/docs/core/index-sparse.html
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

// ?????????????????????
const (
	FTChar FieldType = iota
	FTInt8
	FTInt16
	FTInt32
	FTInt64
	FTUint8
	FTUint16
)

// Column ?????????,??????CreateTable
type Column struct {
	Name          string // ?????????
	Type          string // ??????
	Default       string // ?????????, ??????????????????
	NotNull       bool   // ????????????
	AutoIncrement bool   // ????????????
	PrimaryKey    bool   // ???????????????
	Size          int    // ????????????
}
