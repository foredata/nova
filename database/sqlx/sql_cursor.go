package sqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotPointer      = errors.New("not a pointer")
	ErrNotSlicePointer = errors.New("not slice pointer")
	ErrTooManyColumns  = errors.New("too many columns")
)

func toCursor(rows *sql.Rows, err error) Cursor {
	c := &sqlCursor{rows: rows, err: err}
	if rows != nil && err == nil {
		cols, err := rows.Columns()
		if err != nil {
			c.err = err
		}
		c.columns = cols
	}
	return c
}

type sqlCursor struct {
	rows     *sql.Rows
	err      error
	columns  []string
	pointers []interface{}
}

func (c *sqlCursor) Next() bool {
	return c.rows.Next()
}

func (c *sqlCursor) Decode(out interface{}) error {
	rv := reflect.ValueOf(out)
	if err := isValidPointer(rv); err != nil {
		return err
	}

	return doScan(c.rows, c.columns, getStruct(rv.Type().Elem()), rv.Elem())
}

// All 反射解析多条记录，要求out为slice指针
func (c *sqlCursor) All(out interface{}) error {
	if c.err != nil {
		return c.err
	}

	defer c.rows.Close()
	rv := reflect.ValueOf(out)

	if err := isValidPointer(rv); err != nil {
		return err
	}

	sliceType := rv.Type().Elem()
	if sliceType.Kind() != reflect.Slice {
		return fmt.Errorf("sqlx: %q must be a slice: %w", sliceType.String(), ErrNotSlicePointer)
	}

	sliceVal := reflect.Indirect(reflect.ValueOf(out))
	itemType := sliceType.Elem()
	cols := c.columns

	isPrimitive := itemType.Kind() != reflect.Struct
	if isPrimitive && len(cols) > 1 {
		return ErrTooManyColumns
	}

	info := getStruct(itemType)

	for c.rows.Next() {
		sliceItem := reflect.New(itemType).Elem()
		if err := doScan(c.rows, c.columns, info, sliceItem); err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, sliceItem))
	}

	return c.rows.Err()
}

func (c *sqlCursor) Close() error {
	return c.rows.Close()
}

func (c *sqlCursor) Error() error {
	return c.err
}

func toSingleResult(rows *sql.Rows, err error) SingleResult {
	r := &sqlSingleResult{rows: rows, err: err}
	return r
}

type sqlSingleResult struct {
	rows *sql.Rows
	err  error
}

func (r *sqlSingleResult) Decode(out interface{}) error {
	if r.err != nil {
		return r.err
	}

	rv := reflect.ValueOf(out)
	if err := isValidPointer(rv); err != nil {
		return err
	}

	columns, err := r.rows.Columns()
	if err != nil {
		return fmt.Errorf("sqlx: get columns fail, %w", err)
	}

	return doScan(r.rows, columns, getStruct(rv.Type()), rv.Elem())
}

func (r *sqlSingleResult) Error() error {
	return r.err
}

func isValidPointer(rv reflect.Value) error {
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		return fmt.Errorf("sqlx: target must be a non nil pointer")
	}

	if k := rv.Kind(); k != reflect.Ptr {
		return fmt.Errorf("sqlx: %q must be a pointer: %w", k.String(), ErrNotPointer)
	}

	return nil
}

func doScan(rows *sql.Rows, columns []string, info *dbStruct, out reflect.Value) error {
	switch out.Kind() {
	case reflect.Struct:
		return scanStruct(rows, columns, info, out)
	case reflect.Map:
		return scanMap(rows, columns, out)
	case reflect.Slice:
		return scanSlice(rows, columns, out)
	default:
		return scanPrimitive(rows, out)
	}
}

func initializeNested(v reflect.Value, fieldIndex []int) {
	idx := fieldIndex[0]
	field := v.Field(idx)

	// Create a new instance of a struct and set it to field,
	// if field is a nil pointer to a struct.
	if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct && field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}

	if len(fieldIndex) > 1 {
		initializeNested(reflect.Indirect(field), fieldIndex[1:])
	}
}

func scanStruct(rows *sql.Rows, columns []string, info *dbStruct, out reflect.Value) error {
	scans := make([]interface{}, len(columns))
	for i, col := range columns {
		fieldIndex, ok := info.Fields[col]
		if !ok {
			return fmt.Errorf("sqlx: column: %s not found, or it's unexported in %v", col, out.Type())
		}

		fieldVal := out.FieldByIndex(fieldIndex.Indexes)
		scans[i] = fieldVal.Addr().Interface()
	}

	err := rows.Scan(scans...)
	if err != nil {
		return fmt.Errorf("sqlx: scan struct fail, %w", err)
	}

	return nil
}

func scanMap(rows *sql.Rows, columns []string, out reflect.Value) error {
	if out.IsNil() {
		out.Set(reflect.MakeMap(out.Type()))
	}

	scans := make([]interface{}, len(columns))
	values := make([]reflect.Value, len(columns))

	elemType := out.Type().Elem()
	for i := range columns {
		ptr := reflect.New(elemType)
		scans[i] = ptr.Interface()
		values[i] = ptr.Elem()
	}

	if err := rows.Scan(scans...); err != nil {
		return fmt.Errorf("sqlx: scan rows into map fail, %w", err)
	}

	// We can't set reflect values into destination map before scanning them,
	// because reflect will set a copy, just like regular map behaves,
	// and scan won't modify the map element.
	for i, col := range columns {
		key := reflect.ValueOf(col)
		val := values[i]
		out.SetMapIndex(key, val)
	}

	return nil
}

func scanSlice(rows *sql.Rows, columns []string, out reflect.Value) error {
	scans := make([]interface{}, len(columns))
	values := make([]reflect.Value, len(columns))

	elemType := out.Type().Elem()
	for i := range columns {
		ptr := reflect.New(elemType)
		scans[i] = ptr.Interface()
		values[i] = ptr.Elem()
	}

	if err := rows.Scan(scans...); err != nil {
		return fmt.Errorf("sqlx: scan rows into slice fail, %w", err)
	}

	for _, v := range values {
		out.Set(reflect.Append(out, v))
	}

	return nil
}

func scanPrimitive(rows *sql.Rows, out reflect.Value) error {
	err := rows.Scan(out.Addr().Interface())
	if err != nil {
		return fmt.Errorf("sqlx: scan primitive fail, %w", err)
	}

	return nil
}
