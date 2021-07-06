package sqlx

import (
	"context"
	"database/sql"
	"fmt"
)

func newSqlConn(driver string, dataSourceName string, opts *OpenOptions) (Conn, error) {
	dia := getDialect(opts.Dialect, driver)
	if dia == nil {
		return nil, fmt.Errorf("sql: unknown dialect %q", opts.Dialect)
	}

	return &sqlConn{driver: driver, dsn: dataSourceName, opts: opts, dialect: dia}, nil
}

type sqlConn struct {
	driver  string
	dsn     string
	opts    *OpenOptions
	dialect Dialect
}

func (c *sqlConn) CreateDB(ctx context.Context, name string) error {
	return nil
}

func (c *sqlConn) DropDB(ctx context.Context, name string) error {
	return nil
}

func (c *sqlConn) Database(ctx context.Context, name string) (DB, error) {
	wdb, err := sql.Open(c.driver, c.dsn)
	if err != nil {
		return nil, fmt.Errorf("open writable database fail, %w", err)
	}

	rdb := wdb
	if c.opts.ReadDSN != "" {
		rdb, err = sql.Open(c.driver, c.opts.ReadDSN)
		if err != nil {
			wdb.Close()
			return nil, fmt.Errorf("open readable database fail, %w", err)
		}
	}

	sdb := &sqlDB{name: name, wdb: wdb, rdb: rdb, dialect: c.dialect}

	return sdb, nil
}

func (c *sqlConn) Ping(ctx context.Context) error {
	return nil
}

func (c *sqlConn) Close() error {
	return nil
}
