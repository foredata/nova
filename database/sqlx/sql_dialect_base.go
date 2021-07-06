package sqlx

import (
	"context"
	"database/sql"
)

type sqlResult struct {
	ids  []interface{}
	rows int64
}

func (r *sqlResult) InsertIDs() []interface{} {
	return r.ids
}

func (r *sqlResult) RowsAffected() int64 {
	return r.rows
}

type baseDialect struct {
}

func (b *baseDialect) Indexes(ctx context.Context, db Executor, table string) ([]*Index, error) {
	return nil, errNotSupport
}

// https://www.sqlite.org/lang_createindex.html
func (b *baseDialect) CreateIndex(ctx context.Context, db Executor, table string, index *Index) error {
	sb := sqlBuilder{}

	if index.Unique {
		sb.Write("CREATE UNIQUE INDEX IF NOT EXISTS")
	} else {
		sb.Write("CREATE INDEX IF NOT EXISTS")
	}

	sb.Write(index.toName())
	sb.Write(" ON ", table, " (")
	for _, k := range index.Keys {
		sb.Write(k.Name)
		if k.Order == OrderAsc {
			sb.Write(" AES")
		} else {
			sb.Write(" DESC")
		}
	}

	sb.Write(")")
	return nil
}

func (b *baseDialect) CreateTable(ctx context.Context, db Executor, table string, columns []*Column) error {
	return nil
}

func (b *baseDialect) DropTable(ctx context.Context, db Executor, table string, opts ...*DropTableOptions) error {
	sb := sqlBuilder{}
	sb.Write("DROP TABLE ", table)
	_, err := db.ExecContext(ctx, sb.String())
	return err
}

// INSERT INTO tbl_name (a,b,c) VALUES(1,2,3),(4,5,6),(7,8,9);
// 要求必须同构,以第一个为准
func (b *baseDialect) Insert(ctx context.Context, db Executor, bindType BindVarType, table string, datas []interface{}, opts *InsertOptions) (Result, error) {
	switch len(datas) {
	case 0:
		return &sqlResult{rows: 0}, nil
	case 1:
		return b.InsertOne(ctx, db, bindType, table, datas[0], opts)
	}

	if isValidBindVarType(bindType) {
		return nil, errInvalidBindType
	}

	columns, err := toColumns(datas[0])
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return &sqlResult{}, nil
	}

	fn := bindvarMap[bindType]

	args := make([]interface{}, 0, len(columns)*len(datas))

	sb := sqlBuilder{}
	sb.Write("INSERT INTO ", table, " (")

	for i, c := range columns {
		if i > 0 {
			sb.Write(",")
			sb.Write(",")
		}

		sb.Write(c.Key)
	}
	sb.Write(") VALUES")

	idx := 0
	for i, v := range datas {
		if i > 0 {
			sb.Write(",")
		}
		sb.Write(" (")
		kv, err := toMap(v)
		if err != nil {
			return nil, err
		}

		for _, c := range columns {
			x, ok := kv[c.Key]
			if !ok {
				return nil, errMalformedData
			}
			args = append(args, x)
			sb.Write(fn(idx)...)
			idx++
		}
		sb.Write(")")
	}

	res, err := db.ExecContext(ctx, sb.String(), args...)
	// 注:并不能获取IDs
	return toResult(res, err)
}

// INSERT INTO tbl_name (a,b,c) VALUES(?,?,?);
func (b *baseDialect) InsertOne(ctx context.Context, db Executor, bindType BindVarType, table string, data interface{}, opts *InsertOptions) (Result, error) {
	if isValidBindVarType(bindType) {
		return nil, errInvalidBindType
	}

	columns, err := toColumns(data)
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return &sqlResult{}, nil
	}

	fn := bindvarMap[bindType]

	args := make([]interface{}, 0, len(columns))

	sb := sqlBuilder{}
	sb.Write("INSERT INTO ", table, " (")

	for i, c := range columns {
		if i > 0 {
			sb.Write(",")
		}

		sb.Write(c.Key)
	}

	sb.Write(") VALUES (")
	for i, c := range columns {
		if i > 0 {
			sb.Write(",")
		}
		sb.Write(fn(i)...)
		args = append(args, c.Value)
	}
	sb.Write(")")

	res, err := db.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &sqlResult{ids: []interface{}{id}, rows: 1}, nil
}

// UPDATE table_name SET 列名称 = 新值 WHERE 列名称 = 某值
func (b *baseDialect) Update(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, data interface{}, opts *UpdateOptions) (Result, error) {
	return doUpdate(ctx, db, bindType, table, filter, data, false, opts)
}

// UPDATE Person SET Address = 'Zhongshan 23', City = 'Nanjing' WHERE LastName = 'Wilson' LIMIT 1
func (b *baseDialect) UpdateOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, data interface{}, opts *UpdateOptions) (Result, error) {
	return doUpdate(ctx, db, bindType, table, filter, data, true, opts)
}

func (b *baseDialect) Delete(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *DeleteOptions) (Result, error) {
	if filter == nil {
		sb := sqlBuilder{}
		sb.Write("TRUNCATE TABLE ", table)
		r, err := db.ExecContext(ctx, sb.String())
		return toResult(r, err)
	} else {
		f, err := toFilter(filter)
		if err != nil {
			return nil, err
		}

		sb := sqlBuilder{}
		sb.Write("DELETE FROM ", table, " WHERE ", f.Query())
		r, err := db.ExecContext(ctx, sb.String(), f.Args()...)
		return toResult(r, err)
	}
}

// DELETE FROM table_name WHERE some_column=some_value LIMIT 1;
func (b *baseDialect) DeleteOne(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, opts *DeleteOptions) (Result, error) {
	f, err := toFilter(filter)
	if err != nil {
		return nil, err
	}

	sb := sqlBuilder{}
	sb.Write("DELETE FROM ", table, " WHERE ", f.Query(), " LIMIT 1")
	r, err := db.ExecContext(ctx, sb.String(), f.Args()...)
	return toResult(r, err)
}

func (b *baseDialect) Query(ctx context.Context, db Executor, query string, args ...interface{}) Cursor {
	res, err := db.QueryContext(ctx, query, args...)
	return toCursor(res, err)
}

func (b *baseDialect) Exec(ctx context.Context, db Executor, query string, args ...interface{}) (Result, error) {
	res, err := db.ExecContext(ctx, query, args...)
	return toResult(res, err)
}

func toFilter(f interface{}) (Filter, error) {
	if ff, ok := f.(Filter); ok {
		return ff, nil
	}

	return nil, nil
}

func toResult(r sql.Result, err error) (Result, error) {
	if err != nil {
		return nil, err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return nil, err
	}

	res := &sqlResult{rows: rows}
	return res, nil
}

// UPDATE table_name SET 列名称 = 新值 WHERE 列名称 = 某值
// UPDATE Person SET Address = 'Zhongshan 23', City = 'Nanjing' WHERE LastName = 'Wilson' LIMIT 1
func doUpdate(ctx context.Context, db Executor, bindType BindVarType, table string, filter interface{}, data interface{}, isOne bool, opts *UpdateOptions) (Result, error) {
	if isValidBindVarType(bindType) {
		return nil, errInvalidBindType
	}

	fn := bindvarMap[bindType]

	columns, err := toColumns(data)
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return &sqlResult{}, nil
	}

	f, err := toFilter(filter)
	if err != nil {
		return nil, err
	}

	args := make([]interface{}, 0, len(columns)+2)

	sb := sqlBuilder{}
	sb.Write("UPDATE ", table, " SET ")

	for i, col := range columns {
		if i > 0 {
			sb.Write(",")
		}
		sb.Write(col.Key, "=")
		sb.Write(fn(i)...)

		args = append(args, col.Value)
	}

	// TOOD: bind query
	sb.Write(" WHERE ", f.Query())

	if isOne {
		sb.Write(" LIMIT 1")
	}

	args = append(args, f.Args()...)
	res, err := db.ExecContext(ctx, sb.String(), args...)
	return toResult(res, err)
}

func doSelect(sb *sqlBuilder, table string, f Filter, opts *FindOptions) error {
	sb.Write("SELECT ")
	if len(opts.Fields) == 0 {
		sb.Write("*")
	} else {
		for idx, f := range opts.Fields {
			if idx > 0 {
				sb.Write(",")
			}

			sb.Write(f.Name)
			if f.Alias != "" {
				sb.Write(" AS ", f.Alias)
			}
		}
	}
	sb.Write(" FROM ", table)
	if f != nil {
		sb.Write(" WHEHE ")
		sb.Write(f.Query())
	}

	if len(opts.Sort) > 0 {
		sb.Write(" ORDER BY ")
		for idx, s := range opts.Sort {
			if idx > 0 {
				sb.Write(", ")
			}
			sb.Write(s.Field, " ", s.Order.String())
		}
	}

	return nil
}
