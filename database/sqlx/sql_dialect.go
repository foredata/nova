package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

var (
	errInvalidBindType = errors.New("sqlx: invalid bind type")
	errInvalidData     = errors.New("sqlx: invalid data")
	errMalformedData   = errors.New("sqlx: malformed data")
	errNotSupport      = errors.New("sqlx: not support")
)

func init() {
	RegisterDialect(&mssqlDialect{})
	RegisterDialect(&mysqlDialect{})
	RegisterDialect(&oracleDialect{})
	RegisterDialect(&postgresDialect{})
	RegisterDialect(&sqlite3Dialect{})
}

var (
	dialects       = make(map[string]Dialect)
	dialectDrivers = make(map[string]Dialect)
	dialectsMux    sync.RWMutex
)

// RegisterDialect 注册Dialect
func RegisterDialect(dia Dialect) {
	dialectsMux.Lock()
	dialects[dia.Name()] = dia

	for _, name := range dia.Drivers() {
		dialectDrivers[name] = dia
	}

	dialectsMux.Unlock()
}

func getDialect(name string, rawDriverName string) Dialect {
	dialectsMux.RLock()
	defer dialectsMux.RUnlock()
	var dia Dialect
	if name != "" {
		dia = dialects[name]
	}

	if dia == nil && rawDriverName != "" {
		dia = dialectDrivers[rawDriverName]
	}

	return dia
}

// Executor 标准sql接口
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Dialect 用于实现不同sql之间差异,需要无状态
// [SQL Drivers](https://github.com/golang/go/wiki/SQLDrivers)
// [Go database/sql tutorial](http://go-database-sql.org/index.html)
type Dialect interface {
	// Name 唯一名
	Name() string
	// Drivers 返回已知的Driver名
	Drivers() []string
	// 返回bind variables类型
	BindVar() BindVarType
	//
	Indexes(ctx context.Context, db Executor, table string) ([]*Index, error)
	CreateIndex(ctx context.Context, db Executor, table string, index *Index) error
	DropIndex(ctx context.Context, db Executor, table string, name string) error

	CreateTable(ctx context.Context, db Executor, table string, columns []*Column) error
	DropTable(ctx context.Context, db Executor, table string, opts ...*DropTableOptions) error

	// crud
	Insert(ctx context.Context, db Executor, bindType BindVarType, table string, datas []interface{}, opts *InsertOptions) (Result, error)
	InsertOne(ctx context.Context, db Executor, bindType BindVarType, table string, data interface{}, opts *InsertOptions) (Result, error)
	Update(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, data interface{}, opts *UpdateOptions) (Result, error)
	UpdateOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, data interface{}, opts *UpdateOptions) (Result, error)
	Delete(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *DeleteOptions) (Result, error)
	DeleteOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *DeleteOptions) (Result, error)
	Find(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) Cursor
	FindOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) SingleResult

	// raw sql
	Query(ctx context.Context, db Executor, query string, args ...interface{}) Cursor
	Exec(ctx context.Context, db Executor, query string, args ...interface{}) (Result, error)
}
