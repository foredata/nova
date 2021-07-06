package sqlx

import "context"

type postgresDialect struct {
	baseDialect
}

func (d *postgresDialect) Name() string {
	return "postgres"
}

func (d *postgresDialect) Drivers() []string {
	return []string{"postgres", "pgx", "ramsql"}
}

func (d *postgresDialect) BindVar() BindVarType {
	return BindVarDollar
}

// https://www.postgresql.org/docs/9.3/sql-dropindex.html
func (d *postgresDialect) DropIndex(ctx context.Context, db Executor, table string, name string) error {
	sb := sqlBuilder{}
	sb.Write("DROP INDEX IF EXISTS ", name)
	_, err := db.ExecContext(ctx, sb.String())
	return err
}

// https://www.postgresql.org/docs/8.1/queries-limit.html
func (d *postgresDialect) Find(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) Cursor {
	sb := sqlBuilder{}
	sb.Write("")
	return nil
}

func (d *postgresDialect) FindOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) SingleResult {
	return nil
}
