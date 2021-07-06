package sqlx

// NewFilter 创建Filter
func NewFilter(query string, args ...interface{}) Filter {
	f := &sqlFilter{}
	return f
}

// sqlFilter .
// https://docs.mongodb.com/manual/reference/operator/query/
type sqlFilter struct {
	query string
	args  []interface{}
}

func (f *sqlFilter) Query() string {
	return f.query
}

func (f *sqlFilter) Args() []interface{} {
	return f.args
}
