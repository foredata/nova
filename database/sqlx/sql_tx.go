package sqlx

import (
	"context"
	"database/sql"
)

type sqlTx struct {
	tx      *sql.Tx
	dialect Dialect
}

func (t *sqlTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqlTx) Rollback() error {
	return t.tx.Rollback()
}

func (d *sqlTx) Insert(ctx context.Context, table string, datas []interface{}, opts ...*InsertOptions) (Result, error) {
	return d.dialect.Insert(ctx, d.tx, d.dialect.BindVar(), table, datas, toInsertOptions(opts...))
}

func (d *sqlTx) InsertOne(ctx context.Context, table string, data interface{}, opts ...*InsertOptions) (Result, error) {
	return d.dialect.InsertOne(ctx, d.tx, d.dialect.BindVar(), table, data, toInsertOptions(opts...))
}

func (d *sqlTx) Update(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error) {
	return d.dialect.Update(ctx, d.tx, d.dialect.BindVar(), table, filter, data, toUpdateOptions(opts...))
}

func (d *sqlTx) UpdateOne(ctx context.Context, table string, filter interface{}, data interface{}, opts ...*UpdateOptions) (Result, error) {
	return d.dialect.UpdateOne(ctx, d.tx, d.dialect.BindVar(), table, filter, data, toUpdateOptions(opts...))
}

func (d *sqlTx) Delete(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error) {
	return d.dialect.Delete(ctx, d.tx, d.dialect.BindVar(), table, filter, toDeleteOptions(opts...))
}

func (d *sqlTx) DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*DeleteOptions) (Result, error) {
	return d.dialect.DeleteOne(ctx, d.tx, d.dialect.BindVar(), table, filter, toDeleteOptions(opts...))
}

func (d *sqlTx) Find(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) Cursor {
	return d.dialect.Find(ctx, d.tx, d.dialect.BindVar(), table, filter, toFindOptions(opts...))
}

func (d *sqlTx) FindOne(ctx context.Context, table string, filter interface{}, opts ...*FindOptions) SingleResult {
	return d.dialect.FindOne(ctx, d.tx, d.dialect.BindVar(), table, filter, toFindOptions(opts...))
}

func (d *sqlTx) Query(ctx context.Context, query string, args ...interface{}) Cursor {
	return d.dialect.Query(ctx, d.tx, query, args...)
}

func (d *sqlTx) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return d.dialect.Exec(ctx, d.tx, query, args...)
}
