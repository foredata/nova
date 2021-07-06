package sqlx

import "context"

type mysqlDialect struct {
	baseDialect
}

func (m *mysqlDialect) Name() string {
	return "mysql"
}

func (m *mysqlDialect) Drivers() []string {
	return []string{"mysql", "nrmysql"}
}

func (m *mysqlDialect) BindVar() BindVarType {
	return BindVarQuestion
}

func (m *mysqlDialect) DropIndex(ctx context.Context, db Executor, table string, name string) error {
	sb := sqlBuilder{}
	sb.Write("ALTER TABLE ", table, " DROP INDEX ", name)
	_, err := db.ExecContext(ctx, sb.String())
	return err
}

// Find: [SELECT Statement](https://dev.mysql.com/doc/refman/8.0/en/select.html)
func (m *mysqlDialect) Find(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) Cursor {
	f, err := toFilter(filter)
	if err != nil {
		return toCursor(nil, err)
	}

	sb := sqlBuilder{}
	doSelect(&sb, table, f, opts)

	// limit
	if opts.Limit > 0 {
		sb.Write(" LIMIT ")
		sb.WriteInt(opts.Offset)
		sb.Write(",")
		sb.WriteInt(opts.Limit)
	}

	rows, err := db.QueryContext(ctx, sb.String(), f.Args()...)

	return toCursor(rows, err)
}

func (m *mysqlDialect) FindOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) SingleResult {
	f, err := toFilter(filter)
	if err != nil {
		return toSingleResult(nil, err)
	}

	sb := sqlBuilder{}
	doSelect(&sb, table, f, opts)
	sb.Write(" LIMIT 1")

	rows, err := db.QueryContext(ctx, sb.String(), f.Args()...)

	return toSingleResult(rows, err)
}
