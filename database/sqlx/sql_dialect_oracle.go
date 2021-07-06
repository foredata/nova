package sqlx

import "context"

type oracleDialect struct {
	baseDialect
}

func (d *oracleDialect) Name() string {
	return "oracle"
}

func (d *oracleDialect) Drivers() []string {
	return []string{"oci8", "ora", "goracle", "godror"}
}

func (d *oracleDialect) BindVar() BindVarType {
	return BindVarNamed
}

func (d *oracleDialect) DropIndex(ctx context.Context, db Executor, table string, name string) error {
	sb := sqlBuilder{}
	sb.Write("DROP INDEX ", name)
	_, err := db.ExecContext(ctx, sb.String())
	return err
}

func (m *oracleDialect) Find(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) Cursor {
	sb := sqlBuilder{}
	sb.Write("")
	return nil
}

func (m *oracleDialect) FindOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) SingleResult {
	return nil
}
