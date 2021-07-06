package sqlx

import (
	"context"
	"reflect"
)

// NewScanner .
func NewScanner(ctx context.Context, out interface{}, db DB, table string, filter interface{}, limit int, opts *FindOptions) (Scanner, error) {
	sc := &scanner{
		ctx:    ctx,
		db:     db,
		table:  table,
		filter: filter,
		opts:   opts,
		limit:  limit,
		out:    out,
		outv:   reflect.ValueOf(out),
	}

	return sc, nil
}

// Scanner 区别于Find,Find用于单次查询,而Scanner用于分页遍历全部符合条件的数据并逐条处理
type Scanner interface {
	Next() bool
	Value() interface{}
	Error() error
}

type scanner struct {
	ctx    context.Context
	db     DB
	table  string        //
	filter interface{}   //
	opts   *FindOptions  //
	offset int           // 总偏移
	limit  int           //
	index  int           // 单次偏移
	count  int           // 本次最大值
	out    interface{}   //
	outv   reflect.Value //
	value  interface{}   //
	err    error
}

func (s *scanner) Next() bool {
	if s.index >= s.count {
		s.outv.SetLen(0)
		opt := &FindOptions{
			Offset: s.offset,
			Limit:  s.limit,
		}

		err := s.db.Find(s.ctx, s.table, s.filter, opt).All(s.out)
		if err != nil {
			s.err = err
			return false
		}

		// get count
		s.outv = reflect.ValueOf(s.out)
		count := s.outv.Len()
		s.count = count
		if count == 0 {
			return false
		}
		s.index = 0
		s.move()
		return true
	}

	s.move()
	return true
}

func (s *scanner) move() {
	s.value = s.outv.Index(s.index).Interface()
	s.index++
}

func (s *scanner) Value() interface{} {
	return s.value
}

func (s *scanner) Error() error {
	return s.err
}
