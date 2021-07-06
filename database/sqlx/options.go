package sqlx

import (
	"database/sql"
	"time"
)

type OpenOptions struct {
	ReadDSN         string        // 只读DataSourceName
	MaxIdleConns    int           //
	MaxOpenConns    int           //
	ConnMaxIdleTime time.Duration //
	ConnMaxLifetime time.Duration //
	Dialect         string        //
}

func toOpenOptions(opts ...*OpenOptions) *OpenOptions {
	if len(opts) > 0 {
		return opts[0]
	}

	return nil
}

type DropTableOptions struct {
}

func toDropTableOptions(opts ...*DropTableOptions) *DropTableOptions {
	if len(opts) > 0 {
		return opts[0]
	}

	return nil
}

type InsertOptions struct {
}

func toInsertOptions(opts ...*InsertOptions) *InsertOptions {
	if len(opts) > 0 {
		return opts[0]
	}

	return nil
}

type UpdateOptions struct {
	IgnoreZero bool // 是否忽略空值
}

func toUpdateOptions(opts ...*UpdateOptions) *UpdateOptions {
	if len(opts) > 0 {
		return opts[0]
	}

	return nil
}

type DeleteOptions struct {
}

func toDeleteOptions(opts ...*DeleteOptions) *DeleteOptions {
	if len(opts) > 0 {
		return opts[0]
	}

	return nil
}

type Field struct {
	Name  string // 字段名
	Alias string // 别名
}

type FindOptions struct {
	ReadPref ReadPreference //
	Fields   []Field        // 查询字段,若为空则查询全部
	Sort     []Order        // 排序
	Offset   int
	Limit    int
}

func toFindOptions(opts ...*FindOptions) *FindOptions {
	if len(opts) > 0 {
		return opts[0]
	}

	return nil
}

type TxOptions = sql.TxOptions
