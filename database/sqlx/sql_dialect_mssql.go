package sqlx

import "context"

type mssqlDialect struct {
	baseDialect
}

func (m *mssqlDialect) Name() string {
	return "mssql"
}

func (m *mssqlDialect) Drivers() []string {
	return []string{"sqlserver"}
}

func (m *mssqlDialect) BindVar() BindVarType {
	return BindVarAt
}

func (m *mssqlDialect) DropIndex(ctx context.Context, db Executor, table string, name string) error {
	sb := sqlBuilder{}
	sb.Write("DROP INDEX ", table, ".", name)
	_, err := db.ExecContext(ctx, sb.String())
	return err
}

// support LIMIT: https://blog.csdn.net/sjzs5590/article/details/7337541
func (m *mssqlDialect) Find(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) Cursor {
	return nil
}

func (d *mssqlDialect) FindOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) SingleResult {
	return nil
}
