package sqlx

import (
	"context"
	"database/sql"
	"fmt"
)

type sqlDB struct {
	name    string
	wdb     *sql.DB
	rdb     *sql.DB
	dialect Dialect
}

func (d *sqlDB) Name() string {
	return d.name
}

func (d *sqlDB) Close() error {
	if d.wdb == nil {
		return nil
	}

	if d.wdb == d.rdb {
		db := d.wdb
		d.wdb = nil
		d.rdb = nil
		return db.Close()
	} else {
		werr := d.wdb.Close()
		rerr := d.rdb.Close()
		d.wdb = nil
		d.rdb = nil
		if werr != nil || rerr != nil {
			return fmt.Errorf("sql: close db fail, writeErr=%+v, readErr=%+v", werr, rerr)
		}

		return nil
	}
}

func (d *sqlDB) Indexes(ctx context.Context, table string) ([]*Index, error) {
	return d.dialect.Indexes(ctx, d.wdb, table)
}

func (d *sqlDB) CreateIndex(ctx context.Context, table string, index *Index) error {
	return d.dialect.CreateIndex(ctx, d.wdb, table, index)
}

func (d *sqlDB) DropIndex(ctx context.Context, table string, name string) error {
	return d.dialect.DropIndex(ctx, d.wdb, table, name)
}

func (d *sqlDB) CreateTable(ctx context.Context, table string, columns []*Column) error {
	return d.dialect.CreateTable(ctx, d.wdb, table, columns)
}

func (d *sqlDB) DropTable(ctx context.Context, table string, opts ...*DropTableOptions) error {
	return d.dialect.DropTable(ctx, d.wdb, table, toDropTableOptions(opts...))
}

func (d *sqlDB) Insert(ctx context.Context, table string, datas []interface{}, opts ...*InsertOptions) (Result, error) {
	return d.dialect.Insert(ctx, d.wdb, d.dialect.BindVar(), table, datas, toInsertOptions(opts...))
}

func (d *sqlDB) InsertOne(ctx context.Context, table string, data interface{}, opts ...*InsertOptions) (Result, error) {
	return d.dialect.InsertOne(ctx, d.wdb, d.dialect.BindVar(), table, data, toInsertOptions(opts...))
}

func (d *sqlDB) Update(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error) {
	return d.dialect.Update(ctx, d.wdb, d.dialect.BindVar(), table, filter, data, toUpdateOptions(opts...))
}

func (d *sqlDB) UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error) {
	return d.dialect.UpdateOne(ctx, d.wdb, d.dialect.BindVar(), table, filter, data, toUpdateOptions(opts...))
}

func (d *sqlDB) Delete(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error) {
	return d.dialect.Delete(ctx, d.wdb, d.dialect.BindVar(), table, filter, toDeleteOptions(opts...))
}

func (d *sqlDB) DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error) {
	return d.dialect.DeleteOne(ctx, d.wdb, d.dialect.BindVar(), table, filter, toDeleteOptions(opts...))
}

func (d *sqlDB) Find(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) Cursor {
	return d.dialect.Find(ctx, d.rdb, d.dialect.BindVar(), table, filter, toFindOptions(opts...))
}

func (d *sqlDB) FindOne(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) SingleResult {
	return d.dialect.FindOne(ctx, d.rdb, d.dialect.BindVar(), table, filter, toFindOptions(opts...))
}

func (d *sqlDB) Query(ctx context.Context, query string, args ...interface{}) Cursor {
	return d.dialect.Query(ctx, d.wdb, query, args...)
}

func (d *sqlDB) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return d.dialect.Exec(ctx, d.wdb, query, args...)
}

func (d *sqlDB) Begin(ctx context.Context, opts *TxOptions) (Tx, error) {
	tx, err := d.wdb.BeginTx(ctx, (*sql.TxOptions)(opts))
	if err != nil {
		return nil, err
	}

	return &sqlTx{tx: tx, dialect: d.dialect}, nil
}
