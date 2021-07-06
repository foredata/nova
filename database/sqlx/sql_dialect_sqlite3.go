package sqlx

import "context"

type sqlite3Dialect struct {
	baseDialect
}

func (d *sqlite3Dialect) Name() string {
	return "sqlite3"
}

func (d *sqlite3Dialect) Drivers() []string {
	return nil
}

func (d *sqlite3Dialect) BindVar() BindVarType {
	return BindVarQuestion
}

// https://www.sqlite.org/lang_dropindex.html
func (d *sqlite3Dialect) DropIndex(ctx context.Context, db Executor, table string, name string) error {
	sb := sqlBuilder{}
	sb.Write("DROP INDEX IF EXISTS ", table, ".", name)
	_, err := db.ExecContext(ctx, sb.String())
	return err
}

func (d *sqlite3Dialect) Find(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) Cursor {
	sb := sqlBuilder{}
	sb.Write("")
	return nil
}

func (d *sqlite3Dialect) FindOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *FindOptions) SingleResult {
	return nil
}
